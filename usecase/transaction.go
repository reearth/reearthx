package usecase

import "context"

type Transaction interface {
	Begin() (Tx, error)
}

type Tx interface {
	// Commit informs the Tx to commit when End() is called.
	// If this was not called once, rollback is done when End() is called.
	Commit()
	// End finishes the transaction and do commit if Commit() was called once, or else do rollback.
	// This method is supposed to be called in the uscase layer using defer.
	End(ctx context.Context) error
	// IsCommitted returns true if the Tx is marked as committed.
	IsCommitted() bool
}
