package redisx

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/reearth/reearthx/usecasex"
	"go.uber.org/atomic"
)

type Transaction struct {
	client *redis.Client
}

func NewTransaction(client *redis.Client) *Transaction {
	return &Transaction{
		client: client,
	}
}

func (t *Transaction) Begin(ctx context.Context) (usecasex.Tx, error) {
	pipe := t.client.TxPipeline()
	return &Tx{
		ctx:       ctx,
		pipe:      pipe,
		committed: atomic.NewBool(false),
	}, nil
}

type Tx struct {
	ctx       context.Context
	pipe      redis.Pipeliner
	committed *atomic.Bool
}

func (tx *Tx) Commit() {
	tx.committed.Store(true)
}

func (tx *Tx) IsCommitted() bool {
	return tx.committed.Load()
}

func (tx *Tx) End(ctx context.Context) error {
	defer tx.pipe.Discard()
	
	if !tx.IsCommitted() {
		return nil // Just rollback, no error
	}
	
	// Actually execute the transaction
	_, err := tx.pipe.Exec(ctx)
	return err
}

func (tx *Tx) Context() context.Context {
	return tx.ctx
}
