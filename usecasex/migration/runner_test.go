package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	ctx := context.Background()
	tr := usecasex.NewTransactor(&usecasex.NopTransaction{}, 0)
	r := &dummyRepo{}
	cl := &dummyClient{}
	calls := []Key{}
	m := Migrations[*dummyClient]{
		10: func(ctx context.Context, c *dummyClient) error { calls = append(calls, 10); return nil },
		20: func(ctx context.Context, c *dummyClient) error { calls = append(calls, 20); return nil },
		30: func(ctx context.Context, c *dummyClient) error { calls = append(calls, 30); return nil },
		40: func(ctx context.Context, c *dummyClient) error { calls = append(calls, 40); return nil },
	}
	run := NewRunner[*dummyClient](tr, cl, r, m)

	assert.NoError(t, run.Migrate(ctx))
	assert.Equal(t, 1, r.beginCalls)
	assert.Equal(t, 1, r.endCalls)
	assert.Equal(t, []Key{30, 40}, r.saveCalls) // dummyRepo.Current() == 20
	assert.Equal(t, []Key{30, 40}, calls)
}

func TestRunnerError(t *testing.T) {
	ctx := context.Background()
	tr := usecasex.NewTransactor(&usecasex.NopTransaction{}, 0)
	r := &dummyRepo{}
	calls := []Key{}
	er := errors.New("ERR!")
	m := Migrations[*dummyClient]{
		30: func(ctx context.Context, c *dummyClient) error { calls = append(calls, 30); return er },
	}
	run := NewRunner[*dummyClient](tr, &dummyClient{}, r, m)

	assert.EqualError(t, run.Migrate(ctx), "failed to exec migration 30: ERR!")
	assert.Nil(t, r.saveCalls)
	assert.Equal(t, []Key{30}, calls)
}
