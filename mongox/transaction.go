package mongox

import (
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

func (t *Transaction) Begin() (usecasex.Tx, error) {
	return t.client.BeginTransaction()
}

var _ usecasex.Transaction = (*Transaction)(nil)
