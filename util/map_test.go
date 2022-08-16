package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStableEntries(t *testing.T) {
	assert.Equal(t, []Entry[string, string]{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
	}, StableEntries(map[string]string{
		"b": "2",
		"a": "1",
	}))
}

func TestFromEntries(t *testing.T) {
	assert.Equal(t, map[string]string{
		"a": "1",
		"b": "2",
	}, FromEntries([]Entry[string, string]{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
	}))
}
