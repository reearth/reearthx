package redisx

import (
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/reearth/reearthx/usecasex"
)

type Client struct {
	client      *redis.Client
	transaction usecasex.Transaction
}

func NewClient(opts *redis.Options) *Client {
	client := redis.NewClient(opts)
	return &Client{
		client:      client,
		transaction: &usecasex.NopTransaction{},
	}
}

func NewClientWithClient(client *redis.Client) *Client {
	return &Client{
		client:      client,
		transaction: &usecasex.NopTransaction{},
	}
}

func (c *Client) WithTransaction() *Client {
	c.transaction = NewTransaction(c.client)
	return c
}

func (c *Client) KeySpace(prefix string) *KeySpace {
	return NewKeySpace(c.client, prefix)
}

func (c *Client) WithKeySpace(prefix string) *KeySpace {
	return c.KeySpace(prefix)
}

func (c *Client) Redis() *redis.Client {
	return c.client
}

func (c *Client) Transaction() usecasex.Transaction {
	return c.transaction
}

func IsCluster(redisURL string) bool {
	u, err := url.Parse(redisURL)
	if err != nil {
		return false
	}
	return strings.Contains(u.Host, ",")
}
