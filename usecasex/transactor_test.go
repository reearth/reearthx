package usecasex_test

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactor_RunsAndReturns(t *testing.T) {
	tr := usecasex.NewTransactor(&usecasex.NopTransaction{}, 0)
	ran := false
	err := tr.WithinTransaction(context.Background(), func(ctx context.Context) error {
		ran = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, ran)
}

func TestTransactor_RetriesOnErrTransaction(t *testing.T) {
	tr := usecasex.NewTransactor(&usecasex.NopTransaction{}, 2)
	calls := 0
	err := tr.WithinTransaction(context.Background(), func(ctx context.Context) error {
		calls++
		if calls == 1 {
			return errors.Join(usecasex.ErrTransaction, errors.New("conflict"))
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}
