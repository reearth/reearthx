package pgxx_test

import (
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type kv struct {
	K string
	V int
}

func TestJSONBSlice_RoundTrip(t *testing.T) {
	in := []kv{{"a", 1}, {"b", 2}}
	b, err := pgxx.MarshalJSONBSlice(in)
	require.NoError(t, err)
	out, err := pgxx.UnmarshalJSONBSlice[kv](b)
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestJSONBSlice_EmptyNil(t *testing.T) {
	b, err := pgxx.MarshalJSONBSlice[kv](nil)
	require.NoError(t, err)
	assert.Nil(t, b)

	out, err := pgxx.UnmarshalJSONBSlice[kv](nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}
