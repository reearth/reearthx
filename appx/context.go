package appx

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/log"
)

type requestIDKey struct{}

func ContextMiddleware(key, value any) func(http.Handler) http.Handler {
	return ContextMiddlewareBy(func(w http.ResponseWriter, r *http.Request) context.Context {
		return context.WithValue(r.Context(), key, value)
	})
}

func ContextMiddlewareBy(c func(http.ResponseWriter, *http.Request) context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ctx := c(w, r); ctx == nil {
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return ContextMiddlewareBy(func(w http.ResponseWriter, r *http.Request) context.Context {
		ctx := r.Context()
		reqid := log.GetReqestID(w, r)
		if reqid == "" {
			reqid = uuid.NewString()
		}
		ctx = context.WithValue(ctx, requestIDKey{}, reqid)
		w.Header().Set("X-Request-ID", reqid)

		logger := log.GetLoggerFromContextOrDefault(ctx).SetPrefix(reqid)
		ctx = log.AttachLoggerToContext(ctx, logger)
		return ctx
	})
}

func GetAuthInfo(ctx context.Context, key any) *AuthInfo {
	if ctx == nil {
		return nil
	}
	if auth, ok := ctx.Value(key).(*AuthInfo); ok {
		return auth
	}
	return nil
}

func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqid, ok := ctx.Value(requestIDKey{}).(string); ok {
		return reqid
	}
	return ""
}
