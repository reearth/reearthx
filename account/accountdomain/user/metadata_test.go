package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetadata(t *testing.T) {
	metadata := NewMetadata()
	assert.Equal(t, &Metadata{}, metadata)
}

func TestMetadataFrom(t *testing.T) {
	metadata := MetadataFrom("description", "website")
	assert.Equal(t, "description", metadata.Description())
	assert.Equal(t, "website", metadata.Website())
}

func TestMetadata_Description(t *testing.T) {
	metadata := &Metadata{description: "description"}
	assert.Equal(t, "description", metadata.Description())
}

func TestMetadata_Website(t *testing.T) {
	metadata := &Metadata{website: "website"}
	assert.Equal(t, "website", metadata.Website())
}
