package redisx

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestKeySpaceOperations(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := NewClientWithClient(db)
	
	ctx := context.Background()
	ks := client.KeySpace("test")
	
	mock.ExpectSet("test:key1", "value1", time.Hour).SetVal("OK")
	err := ks.Set(ctx, "key1", "value1", time.Hour).Err()
	assert.NoError(t, err)
	
	mock.ExpectGet("test:key1").SetVal("value1")
	val, err := ks.Get(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
	
	mock.ExpectDel("test:key1").SetVal(1)
	delCount, err := ks.Del(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), delCount)
	
	mock.ExpectLPush("test:list1", "item1", "item2").SetVal(2)
	listLen, err := ks.LPush(ctx, "list1", "item1", "item2").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), listLen)
	
	mock.ExpectLRange("test:list1", 0, -1).SetVal([]string{"item2", "item1"})
	items, err := ks.LRange(ctx, "list1", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"item2", "item1"}, items)
	
	mock.ExpectHSet("test:hash1", "field1", "value1").SetVal(1)
	hsetResult, err := ks.HSet(ctx, "hash1", "field1", "value1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), hsetResult)
	
	mock.ExpectHGet("test:hash1", "field1").SetVal("value1")
	hgetResult, err := ks.HGet(ctx, "hash1", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", hgetResult)
	
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKeySpacePrefixing(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := NewClientWithClient(db)
	
	ctx := context.Background()
	
	prefixes := []string{"users", "products", "orders"}
	
	for _, prefix := range prefixes {
		ks := client.KeySpace(prefix)
		
		mock.ExpectSet(prefix+":key1", "value", time.Minute).SetVal("OK")
		err := ks.Set(ctx, "key1", "value", time.Minute).Err()
		assert.NoError(t, err)
		
		mock.ExpectGet(prefix+":key1").SetVal("value")
		val, err := ks.Get(ctx, "key1").Result()
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	}
	
	assert.NoError(t, mock.ExpectationsWereMet())
}
