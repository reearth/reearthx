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

func Run(m ...Middleware) func(context.Context, Handler) error {
	return func(ctx context.Context, h Handler) (err error) {
		_, err = applyMiddleware(func(ctx context.Context) (context.Context, error) {
			return ctx, h(ctx)
		}, m...)(ctx)
		return
	}
}

func Run1[A any](m ...Middleware) func(context.Context, Handler1[A]) (A, error) {
	return func(ctx context.Context, h Handler1[A]) (a A, err error) {
		_, err = applyMiddleware(func(ctx context.Context) (context.Context, error) {
			a2, err := h(ctx)
			a = a2
			return ctx, err
		}, m...)(ctx)
		return
	}
}

func Run2[A, B any](m ...Middleware) func(context.Context, Handler2[A, B]) (A, B, error) {
	return func(ctx context.Context, h Handler2[A, B]) (a A, b B, err error) {
		_, err = applyMiddleware(func(ctx context.Context) (context.Context, error) {
			a2, b2, err := h(ctx)
			a = a2
			b = b2
			return ctx, err
		}, m...)(ctx)
		return
	}
}

func Run3[A, B, C any](m ...Middleware) func(context.Context, Handler3[A, B, C]) (A, B, C, error) {
	return func(ctx context.Context, h Handler3[A, B, C]) (a A, b B, c C, err error) {
		_, err = applyMiddleware(func(ctx context.Context) (context.Context, error) {
			a2, b2, c2, err := h(ctx)
			a = a2
			b = b2
			c = c2
			return ctx, err
		}, m...)(ctx)
		return
	}
}

func applyMiddleware(h MiddlewareHandler, middleware ...Middleware) MiddlewareHandler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func Context(f func(ctx context.Context) context.Context) Middleware {
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
	return func(next MiddlewareHandler) MiddlewareHandler {
		return func(ctx context.Context) (_ context.Context, err error) {
			tx, err2 := t.Transaction.Begin()
			if err2 != nil {
				return ctx, err2
			}

			defer func() {
				if err2 := tx.End(ctx); err2 != nil && err == nil {
					err = err2
					return
				}
			}()

			ctx2, err := next(ctx)
			if err == nil {
				tx.Commit()
			}
			return ctx2, err
		}
	}
}
