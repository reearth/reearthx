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
		Context(func(ctx context.Context) context.Context {
			ctx = context.WithValue(ctx, ctxkey{}, []string{"1"})
			return ctx
		}),
		Context(func(ctx context.Context) context.Context {
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
	)(
		context.Background(),
		func(ctx context.Context) error {
			assert.Equal(t, []string{"1", "2"}, ctx.Value(ctxkey{}).([]string))
			return nil
		},
	)
	assert.Equal(t, err2, err)

	assert.Same(t, err2, Run()(context.Background(), func(ctx context.Context) error { return err2 }))

	a, err := Run1[int]()(context.Background(), func(ctx context.Context) (int, error) { return 1, err2 })
	assert.Equal(t, 1, a)
	assert.Same(t, err2, err)

	a, b, err := Run2[int, string]()(context.Background(), func(ctx context.Context) (int, string, error) { return 1, "a", err2 })
	assert.Equal(t, 1, a)
	assert.Equal(t, "a", b)
	assert.Same(t, err2, err)

	a, b, c, err := Run3[int, string, bool]()(context.Background(), func(ctx context.Context) (int, string, bool, error) { return 1, "a", true, err2 })
	assert.Equal(t, 1, a)
	assert.Equal(t, "a", b)
	assert.True(t, c)
	assert.Same(t, err2, err)
}

func TestTxUsecase(t *testing.T) {
	// normal
	tr1 := &NopTransaction{}
	uc1 := TxUsecase{Transaction: tr1}
	err1 := Run(uc1.UseTx())(context.Background(), func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err1)
	assert.True(t, tr1.IsCommitted())

	// aborted
	err := errors.New("a")
	tr2 := &NopTransaction{}
	uc2 := TxUsecase{Transaction: tr2}
	err2 := Run(uc2.UseTx())(context.Background(), func(ctx context.Context) error {
		return err
	})
	assert.Same(t, err, err2)
	assert.False(t, tr2.IsCommitted())

	// begin error
	tr3 := &NopTransaction{
		BeginError: err,
	}
	called := false
	uc3 := TxUsecase{Transaction: tr3}
	err3 := Run(uc3.UseTx())(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	assert.Same(t, err, err3)
	assert.False(t, tr3.IsCommitted())
	assert.False(t, called)

	// commit error
	tr4 := &NopTransaction{
		CommitError: err,
	}
	called = false
	uc4 := TxUsecase{Transaction: tr4}
	err4 := Run(uc4.UseTx())(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	assert.Same(t, err, err4)
	assert.True(t, tr4.IsCommitted())
	assert.True(t, called)
}
