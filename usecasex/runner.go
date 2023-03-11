package usecasex

import "context"

type Ctx interface {
	Context() context.Context
	SetContext(ctx context.Context)
}

type Handler func(ctx context.Context) error
type Handler1[A any] func(ctx context.Context) (A, error)
type Handler2[A, B any] func(ctx context.Context) (A, B, error)
type Handler3[A, B, C any] func(ctx context.Context) (A, B, C, error)
type MiddlewareHandler func(ctx context.Context) (context.Context, error)
type Middleware func(next MiddlewareHandler) MiddlewareHandler

func Run(ctx context.Context, h Handler, m ...Middleware) (err error) {
	_, err = ApplyMiddleware(func(ctx context.Context) (context.Context, error) {
		return ctx, h(ctx)
	}, m...)(ctx)
	return
}

func Run1[A any](ctx context.Context, h Handler1[A], m ...Middleware) (a A, err error) {
	_, err = ApplyMiddleware(func(ctx context.Context) (context.Context, error) {
		a2, err := h(ctx)
		a = a2
		return ctx, err
	}, m...)(ctx)
	return
}

func Run2[A, B any](ctx context.Context, h Handler2[A, B], m ...Middleware) (a A, b B, err error) {
	_, err = ApplyMiddleware(func(ctx context.Context) (context.Context, error) {
		a2, b2, err := h(ctx)
		a = a2
		b = b2
		return ctx, err
	}, m...)(ctx)
	return
}

func Run3[A, B, C any](ctx context.Context, h Handler3[A, B, C], m ...Middleware) (a A, b B, c C, err error) {
	_, err = ApplyMiddleware(func(ctx context.Context) (context.Context, error) {
		a2, b2, c2, err := h(ctx)
		a = a2
		b = b2
		c = c2
		return ctx, err
	}, m...)(ctx)
	return
}

func ComposeMiddleware(m ...Middleware) Middleware {
	return func(next MiddlewareHandler) MiddlewareHandler {
		return ApplyMiddleware(next, m...)
	}
}

func ApplyMiddleware(h MiddlewareHandler, middleware ...Middleware) MiddlewareHandler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func UpdateContext(f func(ctx context.Context) context.Context) Middleware {
	return func(next MiddlewareHandler) MiddlewareHandler {
		return func(ctx context.Context) (context.Context, error) {
			return next(f(ctx))
		}
	}
}

type TxUsecase struct {
	Transaction Transaction
}

func (t TxUsecase) UseTx() Middleware {
	return t.UseTxWithRetries(0)
}

func (t TxUsecase) UseTxWithRetries(retry int) Middleware {
	return func(next MiddlewareHandler) MiddlewareHandler {
		return func(ctx context.Context) (ctx2 context.Context, err error) {
			err = DoTransaction(ctx, t.Transaction, retry, func(ctx context.Context) (err2 error) {
				ctx2, err2 = next(ctx)
				return err2
			})
			return
		}
	}
}
