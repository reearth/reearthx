package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopTransactional(t *testing.T) {
	err := errors.New("!")
	tr := &NopTransaction{}

	tr.BeginError = err
	tx, gotErr := tr.Begin()
	assert.Nil(t, tx)
	assert.Same(t, err, gotErr)

	tr.BeginError = nil
	tx, gotErr = tr.Begin()
	assert.NoError(t, gotErr)
	assert.False(t, tr.IsCommitted())

	tx.Commit()
	assert.True(t, tr.IsCommitted())

	tr.CommitError = err
	assert.Same(t, err, tx.End(context.Background()))

	tr.CommitError = nil
	assert.Nil(t, tx.End(context.Background()))
}
