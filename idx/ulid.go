package idx

import (
	"errors"
	"io"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/oklog/ulid"
	"github.com/samber/lo"
)

// ErrInvalidID represents an error for an invalid ID format.
// entropyPool is a sync.Pool providing ULID entropy sources with a monotonic generator.
var (
	ErrInvalidID = errors.New("invalid ID format")

	entropyPool = sync.Pool{
		New: func() interface{} {
			return ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		},
	}
)

// generateID creates a new deterministic ULID using a monotonic source to ensure uniqueness under a given timestamp.
// It retrieves an entropy reader from a sync.Pool for performance optimization and releases it back after usage.
func generateID() ulid.ULID {
	entropy := entropyPool.Get().(io.Reader)
	defer entropyPool.Put(entropy)
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy)
}

// BatchResult represents the result of a batch operation, containing a unique identifier and a potential error.
type BatchResult struct {
	ID  ulid.ULID
	Err error
}

// generateAllID generates a slice of `n` unique ULIDs using concurrent workers for faster processing.
// The function divides the task among available CPU cores to optimize parallelization.
// Returns nil if `n` is less than or equal to zero.
func generateAllID(n int) []ulid.ULID {
	if n <= 0 {
		return nil
	}

	workers := runtime.NumCPU()
	if workers > n {
		workers = n
	}

	ids := make([]ulid.ULID, n)
	idsChan := make(chan BatchResult, n)

	batchSize := n / workers
	remainder := n % workers

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(startIdx, count int) {
			defer wg.Done()

			entropy := entropyPool.Get().(io.Reader)
			defer entropyPool.Put(entropy)

			now := ulid.Timestamp(time.Now().UTC())
			for j := 0; j < count; j++ {
				id := ulid.MustNew(now, entropy)
				idsChan <- BatchResult{ID: id}
			}
		}(i*batchSize, batchSize+(map[bool]int{true: 1, false: 0}[i == workers-1]*remainder))
	}

	go func() {
		wg.Wait()
		close(idsChan)
	}()

	for i := 0; i < n; i++ {
		result := <-idsChan
		if result.Err != nil {
			continue
		}
		ids[i] = result.ID
	}

	return ids
}

// parseID validates and parses a given string into a ULID, returning an error if the format or casing is invalid.
func parseID(id string) (parsedID ulid.ULID, err error) {
	if len(id) != 26 {
		return parsedID, ErrInvalidID
	}

	if strings.IndexFunc(id, unicode.IsUpper) != -1 {
		return parsedID, ErrInvalidID
	}

	return ulid.Parse(id)
}

// mustParseID parses the provided ID string into a ULID and panics if the ID is invalid or cannot be parsed.
func mustParseID(id string) ulid.ULID {
	return lo.Must(parseID(id))
}

// IsValidID checks if the provided ID string has a valid format by attempting to parse it and returns true if valid.
func IsValidID(id string) bool {
	_, err := parseID(id)
	return err == nil
}

// FormatID converts a ULID to its string representation in lowercase.
func FormatID(id ulid.ULID) string {
	return strings.ToLower(id.String())
}
