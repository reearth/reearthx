package redisx

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type KeySpace struct {
	client *redis.Client
	prefix string
}

func NewKeySpace(client *redis.Client, prefix string) *KeySpace {
	return &KeySpace{
		client: client,
		prefix: prefix,
	}
}

func (ks *KeySpace) prefixKey(key string) string {
	return ks.prefix + ":" + key
}

// Redis operations with prefixed keys

func (ks *KeySpace) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return ks.client.Set(ctx, ks.prefixKey(key), value, expiration)
}

func (ks *KeySpace) Get(ctx context.Context, key string) *redis.StringCmd {
	return ks.client.Get(ctx, ks.prefixKey(key))
}

func (ks *KeySpace) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = ks.prefixKey(key)
	}
	return ks.client.Del(ctx, prefixedKeys...)
}

func (ks *KeySpace) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return ks.client.LPush(ctx, ks.prefixKey(key), values...)
}

func (ks *KeySpace) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return ks.client.RPush(ctx, ks.prefixKey(key), values...)
}

func (ks *KeySpace) LPop(ctx context.Context, key string) *redis.StringCmd {
	return ks.client.LPop(ctx, ks.prefixKey(key))
}

func (ks *KeySpace) RPop(ctx context.Context, key string) *redis.StringCmd {
	return ks.client.RPop(ctx, ks.prefixKey(key))
}

func (ks *KeySpace) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return ks.client.LRange(ctx, ks.prefixKey(key), start, stop)
}

func (ks *KeySpace) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return ks.client.HSet(ctx, ks.prefixKey(key), values...)
}

func (ks *KeySpace) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return ks.client.HGet(ctx, ks.prefixKey(key), field)
}

func (ks *KeySpace) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return ks.client.HDel(ctx, ks.prefixKey(key), fields...)
}

func (ks *KeySpace) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return ks.client.Expire(ctx, ks.prefixKey(key), expiration)
}
