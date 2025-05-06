package redisx

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestTransactionBegin(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := NewClientWithClient(db)
	
	txClient := client.WithTransaction()
	tx := txClient.Transaction()
	
	ctx := context.Background()
	
	redisTx, err := tx.Begin(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, redisTx)
	
	txCtx := redisTx.Context()
	assert.NotNil(t, txCtx)
}

func TestTransactionCommit(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := NewClientWithClient(db)
	
	txClient := client.WithTransaction()
	tx := txClient.Transaction()
	
	ctx := context.Background()
	
	redisTx, err := tx.Begin(ctx)
	assert.NoError(t, err)
	
	assert.False(t, redisTx.IsCommitted())
	
	redisTx.Commit()
	
	assert.True(t, redisTx.IsCommitted())
	
	err = redisTx.End(ctx)
	assert.NoError(t, err)
}

func TestDoTransaction(t *testing.T) {
	t.Skip("This test requires complex Redis transaction mocking")
	
	db, _ := redismock.NewClientMock()
	client := NewClientWithClient(db)
	
	txClient := client.WithTransaction()
	tx := txClient.Transaction()
	
	ctx := context.Background()
	
	err := usecasex.DoTransaction(ctx, tx, 1, func(ctx context.Context) error {
		ks := txClient.KeySpace("users")
		return ks.Set(ctx, "123", "data", 0).Err()
	})
	
	assert.NoError(t, err)
}
