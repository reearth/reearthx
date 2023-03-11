package mongox

import (
	"context"
	"errors"

	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

type Transaction struct {
	client *mongo.Client
}

var _ usecasex.Transaction = (*Transaction)(nil)

func NewTransaction(client *mongo.Client) *Transaction {
	return &Transaction{
		client: client,
	}
}

func (t *Transaction) Begin(ctx context.Context) (usecasex.Tx, error) {
	s, err := t.client.StartSession(options.Session())
	if err != nil {
		return nil, err
	}

	if err := s.StartTransaction(options.Transaction()); err != nil {
		return nil, err
	}

	return newTx(ctx, s), nil
}

func IsTransactionError(err error) bool {
	return errorHasLabel(err, driver.TransientTransactionError)
}

// errorHasLabel returns true if err contains the specified label
func errorHasLabel(err error, label string) bool {
	for ; err != nil; err = errors.Unwrap(err) {
		if le, ok := err.(labeledError); ok && le.HasErrorLabel(label) {
			return true
		}
	}
	return false
}

type labeledError interface {
	// HasErrorLabel returns true if the error contains the specified label.
	HasErrorLabel(string) bool
}
