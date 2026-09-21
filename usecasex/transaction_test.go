package usecasex

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ Transaction = (*NopTransaction)(nil)
var _ Tx = (*NopTx)(nil)

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

	assert.Equal(t, ctx, tx.Context())

	tx.Commit()
	assert.True(t, tr.IsCommitted())

	tr.CommitError = err
	assert.Same(t, err, tx.End(context.Background()))

	tr.CommitError = nil
	assert.Nil(t, tx.End(context.Background()))
}

func TestDoTransaction(t *testing.T) {
	ctx := context.Background()
	tr := &NopTransaction{}
	r := 0

	err := DoTransaction(ctx, tr, 2, func(ctx context.Context) error {
		r++
		if r <= 2 {
			return ErrTransaction
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, r)

	r = 0
	err = DoTransaction(ctx, tr, 2, func(ctx context.Context) error {
		r++
		return ErrTransaction
	})
	assert.Same(t, ErrTransaction, err)
	assert.Equal(t, 3, r)

	r = 0
	err = DoTransaction(ctx, tr, -1, func(ctx context.Context) error {
		r++
		return ErrTransaction
	})
	assert.Same(t, ErrTransaction, err)
	assert.Equal(t, 1, r)

	r = 0
	cerr := errors.New("commit")
	tr.CommitError = cerr
	err = DoTransaction(ctx, tr, 3, func(ctx context.Context) error {
		r++
		return nil
	})
	assert.Same(t, cerr, err)
	assert.Equal(t, 1, r)

	r = 0
	tr.CommitError = nil
	err = DoTransaction(ctx, nil, 3, func(ctx context.Context) error {
		r++
		return ErrTransaction
	})
	assert.Same(t, ErrTransaction, err)
	assert.Equal(t, 1, r)
}

func TestDoTransaction_RetryBeginsANewTransaction(t *testing.T) {
	tr := &countingTransaction{}
	calls := 0

	err := DoTransaction(context.Background(), tr, 2, func(context.Context) error {
		calls++
		return ErrTransaction
	})

	assert.Same(t, ErrTransaction, err)
	assert.Equal(t, 3, calls)
	assert.Equal(t, 3, tr.begins)
}

type countingTransaction struct {
	NopTransaction
	begins int
}

func (t *countingTransaction) Begin(ctx context.Context) (Tx, error) {
	t.begins++
	return t.NopTransaction.Begin(ctx)
}
