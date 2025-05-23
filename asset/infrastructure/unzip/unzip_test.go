package unzip

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAsyncUnzipper_Unzip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "unzip_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a zip file for testing
	zipFilePath := filepath.Join(tempDir, "test.zip")
	err = createTestZip(zipFilePath)
	if err != nil {
		t.Fatalf("Failed to create test zip file: %v", err)
	}

	// Set unzip options
	options := UnzipOptions{
		DestPath:      tempDir,
		BufferSize:    1024,
		MaxGoroutines: 2,
		SkipExisting:  false,
	}

	// Create AsyncUnzipper instance
	unzipper := NewAsyncUnzipper(options)

	// Perform unzip
	progressChan, err := unzipper.Unzip(zipFilePath)
	if err != nil {
		t.Fatalf("Failed to start unzip: %v", err)
	}

	// Read progress
	for progress := range progressChan {
		if progress.Error != nil {
			t.Errorf("Error unzipping file %s: %v", progress.CurrentFile, progress.Error)
		}
		t.Logf("Unzipped file: %s (%d/%d)", progress.CurrentFile, progress.CompletedFiles, progress.TotalFiles)
	}

	// Verify unzip result
	expectedFiles := []string{"file1.txt", "file2.txt"}
	for _, fileName := range expectedFiles {
		filePath := filepath.Join(tempDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be unzipped, but it was not found", fileName)
		}
	}
}

// createTestZip creates a zip file with two files for testing
func createTestZip(zipFilePath string) error {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	files := []struct {
		Name, Body string
	}{
		{"file1.txt", "This is the content of file1."},
		{"file2.txt", "This is the content of file2."},
	}

	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			return err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(zipFilePath, buf.Bytes(), 0644)
}
