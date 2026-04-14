package log

import (
	"fmt"
	"time"

	echov5 "github.com/labstack/echo/v5"
)

// AccessLoggerV5 returns an echo v5 middleware that logs request/response pairs.
func AccessLoggerV5(l *Logger) echov5.MiddlewareFunc {
	return func(next echov5.HandlerFunc) echov5.HandlerFunc {
		return func(c *echov5.Context) error {
			req := c.Request()
			res, _ := echov5.UnwrapResponse(c.Response())
			start := time.Now()

			reqid := GetReqestID(res, req)
			args := []any{
				"time_unix", start.Unix(),
				"remote_ip", c.RealIP(),
				"host", req.Host,
				"uri", req.RequestURI,
				"method", req.Method,
				"path", req.URL.Path,
				"protocol", req.Proto,
				"referer", req.Referer(),
				"user_agent", req.UserAgent(),
				"bytes_in", req.ContentLength,
				"request_id", reqid,
				"route", c.Path(),
			}

			logger := GetLoggerFromContext(req.Context())
			if logger == nil {
				logger = l
			}
			logger = logger.WithCaller(false)

			// incoming log
			logger.Infow(
				fmt.Sprintf("<-- %s %s", req.Method, req.URL.Path),
				args...,
			)

			err := next(c)
			if err != nil {
				c.Echo().HTTPErrorHandler(c, err)
			}

			resp, status := echov5.ResolveResponseStatus(c.Response(), err)
			stop := time.Now()
			latency := stop.Sub(start)
			latencyHuman := latency.String()

			args = append(args,
				"status", status,
				"bytes_out", resp.Size,
				"latency", latency.Microseconds(),
				"latency_human", latencyHuman,
			)

			// outcoming log
			logger.Infow(
				fmt.Sprintf("--> %s %d %s %s", req.Method, status, req.URL.Path, latencyHuman),
				args...,
			)
			return nil
		}
	}
}
