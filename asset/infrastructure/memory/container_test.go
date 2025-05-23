package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	got := New()
	assert.NotNil(t, got)
	assert.NotNil(t, got.Asset)
	assert.NotNil(t, got.AssetFile)
	assert.NotNil(t, got.Project)
	assert.NotNil(t, got.Thread)
	assert.NotNil(t, got.Event)
	assert.NotNil(t, got.Transaction)
}
