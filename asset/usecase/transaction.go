package assetusecase

import (
	"context"
)

// TransactionManager defines the interface for managing transactions
type TransactionManager interface {
	// WithTransaction executes the given function within a transaction
	// If the function returns an error, the transaction is rolled back
	// If the function returns nil, the transaction is committed
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionKey is the context key for storing transaction information
type transactionKey struct{}

// NewTransactionContext creates a new context with transaction information
func NewTransactionContext(ctx context.Context, tx interface{}) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

// GetTransactionFromContext retrieves transaction information from context
func GetTransactionFromContext(ctx context.Context) interface{} {
	return ctx.Value(transactionKey{})
}
