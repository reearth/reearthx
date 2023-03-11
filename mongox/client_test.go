package mongox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTransactionAvailable(t *testing.T) {
	assert.False(t, IsTransactionAvailable("mongodb://localhost"))
	assert.True(t, IsTransactionAvailable("mongodb://localhost,localhost2"))
	assert.True(t, IsTransactionAvailable("mongodb+srv://xxx:xxx@xxx.example.com"))
}
