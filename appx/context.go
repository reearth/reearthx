package appx

import (
	"context"
	"net/http"
	"os"

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
	googleCloudProject := os.Getenv("GOOGLE_CLOUD_PROJECT")

	return ContextMiddlewareBy(func(w http.ResponseWriter, r *http.Request) context.Context {
		ctx := r.Context()
		reqid := log.GetReqestID(w, r)
		if reqid == "" {
			reqid = uuid.NewString()
		}
		ctx = context.WithValue(ctx, requestIDKey{}, reqid)
		logger := log.GetLoggerFromContextOrDefault(ctx).SetPrefix(reqid)

		// https://cloud.google.com/run/docs/logging#correlate-logs
		if googleTrace := log.GoogleTraceFromTraceID(
			log.TraceIDFrom(reqid),
			googleCloudProject,
		); googleTrace != "" {
			logger = logger.With("logging.googleapis.com/trace", googleTrace)
		} else {
			w.Header().Set("X-Request-ID", reqid)
		}

		ctx = log.ContextWith(ctx, logger)
		return ctx
	})
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
