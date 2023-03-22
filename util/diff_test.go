package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	// case1
	assert.Equal(t, DiffResult[string]{
		Added: []string{"a|1"},
		Updated: []DiffUpdate[string]{
			{Prev: "b|2", Next: "b|3"},
			{Prev: "d|4", Next: "d|5"},
		},
		NotUpdated: []DiffUpdate[string]{
			{Prev: "c|3", Next: "c|3"},
		},
		Deleted: []string{"e|5"},
	}, Diff(
		[]string{"b|2", "c|3", "d|4", "e|5"},
		[]string{"a|1", "b|3", "c|3", "d|5"},
		func(a, b string) bool {
			ap, _, _ := strings.Cut(a, "|")
			bp, _, _ := strings.Cut(b, "|")
			return ap == bp
		},
		func(a, b string) bool {
			_, as, _ := strings.Cut(a, "|")
			_, bs, _ := strings.Cut(b, "|")
			return as != bs
		},
	))

	// case2
	assert.Equal(t, DiffResult[string]{
		Deleted: []string{"b|2", "c|3", "d|4"},
	}, Diff(
		[]string{"b|2", "c|3", "d|4"},
		nil,
		func(a, b string) bool {
			ap, _, _ := strings.Cut(a, "|")
			bp, _, _ := strings.Cut(b, "|")
			return ap == bp
		},
		func(a, b string) bool {
			_, as, _ := strings.Cut(a, "|")
			_, bs, _ := strings.Cut(b, "|")
			return as != bs
		},
	))

	r := DiffResult[int]{
		Updated: []DiffUpdate[int]{
			{Prev: 1, Next: 2}, {Prev: 3, Next: 4},
		},
		NotUpdated: []DiffUpdate[int]{
			{Prev: 5, Next: 5},
		},
	}

	assert.Equal(t, []int{5}, r.NotUpdatedNext())
	assert.Equal(t, []int{5}, r.NotUpdatedPrev())
	assert.Equal(t, []int{2, 4}, r.UpdatedNext())
	assert.Equal(t, []int{1, 3}, r.UpdatedPrev())
}
