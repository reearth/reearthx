package usecasex

import "context"

// Transactor runs a function within a database transaction, committing on a nil
// return and rolling back on error. The fn receives a context carrying the
// transaction; repositories resolve it (e.g. via pgxx.Executor) transparently.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// NewTransactor adapts an existing Transaction (e.g. the Mongo implementation)
// into a Transactor, delegating to DoTransaction so retry-on-ErrTransaction
// behavior is preserved. retry <= 0 means a single attempt.
func NewTransactor(t Transaction, retry int) Transactor {
	return &transactorAdapter{t: t, retry: retry}
}

type transactorAdapter struct {
	t     Transaction
	retry int
}

func (a *transactorAdapter) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return DoTransaction(ctx, a.t, a.retry, fn)
}
