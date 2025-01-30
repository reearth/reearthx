package idx

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/samber/lo"
)

var entropyPool = sync.Pool{
	New: func() interface{} {
		return ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	},
}

func generateID() ulid.ULID {
	entropy := entropyPool.Get().(io.Reader)
	defer entropyPool.Put(entropy)
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy)
}

func generateAllID(n int) []ulid.ULID {
	ids := make([]ulid.ULID, 0, n)
	entropy := entropyPool.Get().(io.Reader)
	defer entropyPool.Put(entropy)
	for i := 0; i < n; i++ {
		newID := ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy)
		ids = append(ids, newID)
	}
	return ids
}

func parseID(id string) (parsedID ulid.ULID, e error) {
	if includeUpperCase(id) {
		return parsedID, ErrInvalidID
	}
	return ulid.Parse(id)
}

func includeUpperCase(s string) bool {
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			return true
		}
	}
	return false
}

func mustParseID(id string) ulid.ULID {
	return lo.Must(parseID(id))
}
