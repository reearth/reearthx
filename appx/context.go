package appx

import (
	"context"
	"net/http"
)

func ContextMiddleware(key, value any) func(http.Handler) http.Handler {
	return ContextMiddlewareBy(func(r *http.Request) context.Context {
		return context.WithValue(r.Context(), key, value)
	})
}

func ContextMiddlewareBy(c func(*http.Request) context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ctx := c(r); ctx == nil {
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}
