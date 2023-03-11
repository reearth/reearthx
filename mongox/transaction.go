package mongox

import (
	"context"

	"github.com/reearth/reearthx/usecasex"
)

type Transaction struct {
	client *Client
}

func NewTransaction(client *Client) *Transaction {
	return &Transaction{
		client: client,
	}
}

func (t *Transaction) Begin(ctx context.Context) (usecasex.Tx, error) {
	return t.client.BeginTransaction(ctx)
}

var _ usecasex.Transaction = (*Transaction)(nil)
