package usecasex

import (
	"context"

	"go.uber.org/atomic"
)

type Transaction interface {
	Begin(context.Context) (Tx, error)
}

type Tx interface {
	// Context returns a context suitable for use under transaction.
	Context() context.Context
	// Commit informs the Tx to commit when End() is called.
	// If this was not called once, rollback is done when End() is called.
	Commit()
	// End finishes the transaction and do commit if Commit() was called once, or else do rollback.
	// This method is supposed to be called in the uscase layer using defer.
	End(ctx context.Context) error
	// IsCommitted returns true if the Tx is marked as committed.
	IsCommitted() bool
}

type NopTransaction struct {
	BeginError  error
	CommitError error
	committed   atomic.Bool
}

type NopTx struct {
	ctx context.Context
	t   *NopTransaction
}

func (t *NopTransaction) Begin(ctx context.Context) (Tx, error) {
	if t.BeginError != nil {
		return nil, t.BeginError
	}
	return &NopTx{ctx: ctx, t: t}, nil
}

func (t *NopTransaction) IsCommitted() bool {
	return t.committed.Load()
}

func (t *NopTx) Commit() {
	t.t.committed.Store(true)
}

func (t *NopTx) IsCommitted() bool {
	return t.t.committed.Load()
}

func (t *NopTx) End(_ context.Context) error {
	return t.t.CommitError
}

func (t *NopTx) Context() context.Context {
	return t.ctx
}

var _ Transaction = (*NopTransaction)(nil)
var _ Tx = (*NopTx)(nil)
