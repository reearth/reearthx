package mongox

import "github.com/reearth/reearthx/usecase"

type Transaction struct {
	client *Client
}

func NewTransaction(client *Client) *Transaction {
	return &Transaction{
		client: client,
	}
}

func (t *Transaction) Begin() (usecase.Tx, error) {
	return t.client.BeginTransaction()
}

var _ usecase.Transaction = (*Transaction)(nil)
