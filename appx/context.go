package appx

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/log"
)

type requestIDKey struct{}

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

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return ContextMiddlewareBy(func(r *http.Request) context.Context {
		ctx := r.Context()
		reqid := getHeader(r,
			"X-Request-ID",
			"X-Amzn-Trace-Id",       // AWS
			"X-Cloud-Trace-Context", // GCP
			"X-ARR-LOG-ID",          // Azure
		)
		if reqid == "" {
			reqid = uuid.NewString()
		}
		ctx = context.WithValue(ctx, requestIDKey{}, reqid)

		logger := log.GetLoggerFromContextOrDefault(ctx).SetPrefix(reqid)
		ctx = log.AttachLoggerToContext(ctx, logger)
		return ctx
	})
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqid, ok := ctx.Value(requestIDKey{}).(string); ok {
		return reqid
	}
	return ""
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

func getHeader(r *http.Request, keys ...string) string {
	for _, k := range keys {
		if v := r.Header.Get(k); v != "" {
			return v
		}
		if v := r.Header.Get(strings.ToLower(k)); v != "" {
			return v
		}
	}
	return ""
}
