package usecasex

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopTransactional(t *testing.T) {
	err := errors.New("!")
	tr := &NopTransaction{}
	ctx := context.Background()

	tr.BeginError = err
	tx, gotErr := tr.Begin(ctx)
	assert.Nil(t, tx)
	assert.Same(t, err, gotErr)

	tr.BeginError = nil
	tx, gotErr = tr.Begin(ctx)
	assert.NoError(t, gotErr)
	assert.False(t, tr.IsCommitted())

	assert.Same(t, ctx, tx.Context())

	tx.Commit()
	assert.True(t, tr.IsCommitted())

	tr.CommitError = err
	assert.Same(t, err, tx.End(context.Background()))

	tr.CommitError = nil
	assert.Nil(t, tx.End(context.Background()))
}
