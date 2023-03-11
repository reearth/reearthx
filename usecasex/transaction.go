package usecasex

import (
	"context"
	"errors"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"go.uber.org/atomic"
)

const IDErrTransaction = "transaction error"

var (
	ErrTransaction    = rerror.WrapE(&i18n.Message{ID: IDErrTransaction}, rawErrTransaction)
	rawErrTransaction = errors.New("transaction conflicted")
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

func DoTransaction(ctx context.Context, t Transaction, retry int, fn func(ctx context.Context) error) (err error) {
	if t == nil {
		return fn(ctx)
	}

	tx, err := t.Begin(ctx)
	if err != nil {
		return err
	}

	ctx2 := tx.Context()
	defer func() {
		if err2 := tx.End(ctx2); err2 != nil && err == nil {
			err = err2
		}
	}()

	r := 0
	for {
		if r > 0 && (retry <= 0 || r > retry) {
			break
		}
		if err = fn(ctx2); err != nil {
			if !errors.Is(err, ErrTransaction) {
				break
			}
		} else {
			tx.Commit()
			break
		}
		r++
	}

	return err
}
