package usecasex

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	type ctxkey struct{}
	err2 := errors.New("a")
	err := Run(
		context.Background(),
		func(ctx context.Context) error {
			assert.Equal(t, []string{"1", "2"}, ctx.Value(ctxkey{}).([]string))
			return nil
		},
		UpdateContext(func(ctx context.Context) context.Context {
			ctx = context.WithValue(ctx, ctxkey{}, []string{"1"})
			return ctx
		}),
		UpdateContext(func(ctx context.Context) context.Context {
			s := ctx.Value(ctxkey{}).([]string)
			s = append(s, "2")
			ctx = context.WithValue(ctx, ctxkey{}, s)
			return ctx
		}),
		func(next MiddlewareHandler) MiddlewareHandler {
			return func(ctx context.Context) (context.Context, error) {
				ctx, err := next(ctx)
				if err == nil {
					return ctx, err2
				}
				return ctx, err
			}
		},
	)
	assert.Equal(t, err2, err)

	assert.Same(t, err2, Run(context.Background(), func(ctx context.Context) error { return err2 }))

	a, err := Run1(context.Background(), func(ctx context.Context) (int, error) { return 1, err2 })
	assert.Equal(t, 1, a)
	assert.Same(t, err2, err)

	a, b, err := Run2(context.Background(), func(ctx context.Context) (int, string, error) { return 1, "a", err2 })
	assert.Equal(t, 1, a)
	assert.Equal(t, "a", b)
	assert.Same(t, err2, err)

	a, b, c, err := Run3(context.Background(), func(ctx context.Context) (int, string, bool, error) { return 1, "a", true, err2 })
	assert.Equal(t, 1, a)
	assert.Equal(t, "a", b)
	assert.True(t, c)
	assert.Same(t, err2, err)
}

func TestTxUsecase(t *testing.T) {
	// normal
	tr1 := &NopTransaction{}
	uc1 := TxUsecase{Transaction: tr1}
	err1 := Run(context.Background(), func(ctx context.Context) error {
		return nil
	}, uc1.UseTx())
	assert.NoError(t, err1)
	assert.True(t, tr1.IsCommitted())

	// aborted
	err := errors.New("a")
	tr2 := &NopTransaction{}
	uc2 := TxUsecase{Transaction: tr2}
	err2 := Run(context.Background(), func(ctx context.Context) error {
		return err
	}, uc2.UseTx())
	assert.Same(t, err, err2)
	assert.False(t, tr2.IsCommitted())

	// begin error
	tr3 := &NopTransaction{
		BeginError: err,
	}
	called := false
	uc3 := TxUsecase{Transaction: tr3}
	err3 := Run(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	}, uc3.UseTx())
	assert.Same(t, err, err3)
	assert.False(t, tr3.IsCommitted())
	assert.False(t, called)

	// commit error
	tr4 := &NopTransaction{
		CommitError: err,
	}
	called = false
	uc4 := TxUsecase{Transaction: tr4}
	err4 := Run(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	}, uc4.UseTx())
	assert.Same(t, err, err4)
	assert.True(t, tr4.IsCommitted())
	assert.True(t, called)
}

func TestComposeMiddleware(t *testing.T) {
	type ctxkey struct{}

	m := ComposeMiddleware(
		UpdateContext(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, ctxkey{}, "a")
		}),
		UpdateContext(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, ctxkey{}, ctx.Value(ctxkey{}).(string)+"b")
		}),
	)

	called := false
	err := Run(context.Background(), func(ctx context.Context) error {
		assert.Equal(t, "ab", ctx.Value(ctxkey{}).(string))
		called = true
		return nil
	}, m)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestApplyMiddleware(t *testing.T) {
	type ctxkey struct{}

	ctx, err := ApplyMiddleware(
		func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		},
		UpdateContext(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, ctxkey{}, "a")
		}),
		UpdateContext(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, ctxkey{}, ctx.Value(ctxkey{}).(string)+"b")
		}),
	)(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "ab", ctx.Value(ctxkey{}).(string))
}
