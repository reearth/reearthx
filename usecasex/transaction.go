package usecasex

import (
	"context"
	"errors"
	"time"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"go.uber.org/atomic"
)

const IDErrTransaction = "transaction error"

var (
	ErrTransaction    = rerror.WrapE(&i18n.Message{ID: IDErrTransaction}, errTransactionConflicted)
	errTransactionConflicted = errors.New("transaction conflicted")
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

// DoTransaction runs fn inside a transaction, retrying up to retry additional
// times on TransientTransactionError (e.g. MongoDB WriteConflict). Each retry
// starts a fresh Begin so the underlying driver can renegotiate locks cleanly,
// which is required for MongoDB multi-document transactions. A linear backoff
// of 50ms × attempt is applied between retries.
func DoTransaction(ctx context.Context, t Transaction, retry int, fn func(ctx context.Context) error) error {
	if t == nil {
		return fn(ctx)
	}

	var lastErr error
	for attempt := 0; attempt == 0 || (retry > 0 && attempt <= retry); attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}

		tx, err := t.Begin(ctx)
		if err != nil {
			return err
		}

		txCtx := tx.Context()
		lastErr = fn(txCtx)
		if lastErr == nil {
			tx.Commit()
		}
		if endErr := tx.End(txCtx); endErr != nil && lastErr == nil {
			lastErr = endErr
		}
		if lastErr == nil {
			return nil
		}
		if !errors.Is(lastErr, ErrTransaction) {
			return lastErr
		}
	}

	return lastErr
}
