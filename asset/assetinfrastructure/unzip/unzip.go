package unzip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// UnzipOptions contains configuration options for the unzip operation
type UnzipOptions struct {
	// DestPath is the destination directory for extracted files
	DestPath string
	// BufferSize is the size of the buffer used for copying files
	BufferSize int
	// MaxGoroutines is the maximum number of concurrent extractions
	MaxGoroutines int
	// SkipExisting determines whether to skip existing files
	SkipExisting bool
}

// Progress represents the progress of the unzip operation
type Progress struct {
	CurrentFile    string
	TotalFiles     int
	CompletedFiles int
	Error          error
}

// AsyncUnzipper handles asynchronous unzip operations
type AsyncUnzipper struct {
	options  UnzipOptions
	progress chan Progress
	wg       sync.WaitGroup
	sem      chan struct{} // semaphore for limiting goroutines
}

// NewAsyncUnzipper creates a new AsyncUnzipper instance
func NewAsyncUnzipper(options UnzipOptions) *AsyncUnzipper {
	if options.BufferSize <= 0 {
		options.BufferSize = 32 * 1024 // 32KB default buffer
	}
	if options.MaxGoroutines <= 0 {
		options.MaxGoroutines = 4 // default concurrent goroutines
	}

	return &AsyncUnzipper{
		options:  options,
		progress: make(chan Progress, options.MaxGoroutines),
		sem:      make(chan struct{}, options.MaxGoroutines),
	}
}

// Unzip starts the async unzip operation
func (au *AsyncUnzipper) Unzip(zipPath string) (<-chan Progress, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}

	totalFiles := len(reader.File)
	completedFiles := 0

	go func() {
		defer reader.Close()
		defer close(au.progress)

		for _, file := range reader.File {
			au.wg.Add(1)
			au.sem <- struct{}{} // acquire semaphore

			go func(f *zip.File) {
				defer au.wg.Done()
				defer func() { <-au.sem }() // release semaphore

				err := au.extractFile(f)
				completedFiles++
				au.progress <- Progress{
					CurrentFile:    f.Name,
					TotalFiles:     totalFiles,
					CompletedFiles: completedFiles,
					Error:          err,
				}
			}(file)
		}

		au.wg.Wait()
	}()

	return au.progress, nil
}

// extractFile extracts a single file from the zip archive
func (au *AsyncUnzipper) extractFile(f *zip.File) error {
	destPath := filepath.Join(au.options.DestPath, f.Name)

	// Create directory for file
	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, f.Mode())
	}

	// Create directory for file if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Check if file exists and skip if configured
	if au.options.SkipExisting {
		if _, err := os.Stat(destPath); err == nil {
			return nil
		}
	}

	// Open the compressed file
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the destination file
	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy the contents
	buf := make([]byte, au.options.BufferSize)
	_, err = io.CopyBuffer(dest, rc, buf)
	return err
}
