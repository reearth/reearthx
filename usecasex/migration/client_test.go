package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	c := &dummyClient{}
	r := &dummyRepo{}
	calls := []Key{}
	m := Migrations[*dummyClient]{
		10: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 10)
			return nil
		},
		20: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 20)
			return nil
		},
		30: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 30)
			return nil
		},
		40: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 40)
			return nil
		},
		50: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 50)
			return nil
		},
	}
	cl := NewClient(c, r, m, 0)

	assert.NoError(t, cl.Migrate(ctx))
	assert.Equal(t, 1, r.beginCalls)
	assert.Equal(t, 1, r.endCalls)
	assert.Equal(t, 1, r.currentCalls)
	assert.Equal(t, []Key{30, 40, 50}, r.saveCalls)
	assert.Equal(t, []Key{30, 40, 50}, calls)
}

func TestClientError(t *testing.T) {
	ctx := context.Background()
	c := &dummyClient{}
	r := &dummyRepo{}
	calls := []Key{}
	err := errors.New("ERR!")
	m := Migrations[*dummyClient]{
		30: func(ctx context.Context, c *dummyClient) error {
			calls = append(calls, 30)
			return err
		},
	}
	cl := NewClient(c, r, m, 0)

	assert.EqualError(t, cl.Migrate(ctx), "Failed to exec migration 30: ERR!")
	assert.Equal(t, 1, r.beginCalls)
	assert.Equal(t, 1, r.endCalls)
	assert.Equal(t, 1, r.currentCalls)
	assert.Nil(t, r.saveCalls)
	assert.Equal(t, []Key{30}, calls)
}

type dummyRepo struct {
	currentCalls int
	saveCalls    []Key
	beginCalls   int
	endCalls     int
}

func (c *dummyRepo) Current(ctx context.Context) (Key, error) {
	c.currentCalls++
	return 20, nil
}

func (c *dummyRepo) Save(ctx context.Context, key Key) error {
	c.saveCalls = append(c.saveCalls, key)
	return nil
}

func (c *dummyRepo) Begin(ctx context.Context) error {
	c.beginCalls++
	return nil
}

func (c *dummyRepo) End(ctx context.Context) error {
	c.endCalls++
	return nil
}

type dummyClient struct{}

func (c *dummyClient) Transaction() usecasex.Transaction {
	return &usecasex.NopTransaction{}
}
