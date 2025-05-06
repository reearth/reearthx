package redisx

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	opts := &redis.Options{
		Addr: "localhost:6379",
	}
	client := NewClient(opts)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Redis())
}

func TestNewClientWithClient(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClientWithClient(redisClient)
	assert.NotNil(t, client)
	assert.Equal(t, redisClient, client.Redis())
}

func TestWithTransaction(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClientWithClient(redisClient)
	txClient := client.WithTransaction()
	assert.NotNil(t, txClient)
	assert.NotNil(t, txClient.Transaction())
}

func TestKeySpace(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClientWithClient(redisClient)
	ks := client.KeySpace("test")
	assert.NotNil(t, ks)
}

func TestWithKeySpace(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClientWithClient(redisClient)
	ks := client.WithKeySpace("test")
	assert.NotNil(t, ks)
}

func TestIsCluster(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "single host",
			url:      "redis://localhost:6379",
			expected: false,
		},
		{
			name:     "multiple hosts",
			url:      "redis://host1:6379,host2:6379,host3:6379",
			expected: true,
		},
		{
			name:     "invalid url",
			url:      "://invalid",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCluster(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}
