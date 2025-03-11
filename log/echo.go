package log

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/reearth/reearthx/util"
	"go.uber.org/zap/zapcore"
)

type Echo struct {
	logger *Logger
}

var _ echo.Logger = (*Echo)(nil)

func NewEcho() *Echo {
	return &Echo{
		logger: globalLogger,
	}
}

func NewEchoWith(logger *Logger) *Echo {
	return &Echo{
		logger: logger.AddCallerSkip(1),
	}
}

func NewEchoWithRaw(logger *Logger) *Echo {
	return &Echo{
		logger: logger,
	}
}

func (l *Echo) SetDynamicPrefix(prefix func() Format) {
	l.logger = l.logger.SetDynamicPrefix(prefix)
}

func (l *Echo) SetDynamicSuffix(suffix func() Format) {
	l.logger = l.logger.SetDynamicSuffix(suffix)
}

// Level returns logger level
func (l *Echo) Level() log.Lvl {
	switch l.logger.Level() {
	case zapcore.DebugLevel:
		return log.DEBUG
	case zapcore.InfoLevel:
		return log.INFO
	case zapcore.WarnLevel:
		return log.WARN
	case zapcore.ErrorLevel:
		return log.ERROR
	default:
		l.Panic("Invalid level")
	}
	return log.OFF
}

// SetHeader is a stub to satisfy interface
// It's controlled by Logger
func (l *Echo) SetHeader(_ string) {}

// SetPrefix It's controlled by Logger
func (l *Echo) SetPrefix(s string) {
	l.logger = l.logger.SetPrefix(s)
}

// Prefix It's controlled by Logger
func (l *Echo) Prefix() string {
	return l.logger.Prefix()
}

// SetLevel set level to logger from given log.Lvl
func (l *Echo) SetLevel(lvl log.Lvl) {
	switch lvl {
	case log.DEBUG:
		l.logger.SetLevel(zapcore.DebugLevel)
	case log.INFO:
		l.logger.SetLevel(zapcore.InfoLevel)
	case log.WARN:
		l.logger.SetLevel(zapcore.WarnLevel)
	case log.ERROR:
		l.logger.SetLevel(zapcore.ErrorLevel)
	}
}

// Output logger output func
func (l *Echo) Output() io.Writer {
	return DefaultOutput
}

// SetOutput change output, default os.Stdout
func (l *Echo) SetOutput(w io.Writer) {
	l.logger = l.logger.SetOutput(w)
}

// Printj print JSON log
func (l *Echo) Printj(j log.JSON) {
	l.logger.Print(fromMap(j))
}

// Debugj debug JSON log
func (l *Echo) Debugj(j log.JSON) {
	l.logger.Debug(fromMap(j))
}

// Infoj info JSON log
func (l *Echo) Infoj(j log.JSON) {
	l.logger.Info(fromMap(j))
}

// Warnj warning JSON log
func (l *Echo) Warnj(j log.JSON) {
	l.logger.Warn(fromMap(j))
}

// Errorj error JSON log
func (l *Echo) Errorj(j log.JSON) {
	l.logger.Error(fromMap(j))
}

// Fatalj fatal JSON log
func (l *Echo) Fatalj(j log.JSON) {
	l.logger.Fatal()
}

// Panicj panic JSON log
func (l *Echo) Panicj(j log.JSON) {
	l.logger.Panic()
}

// Print string log
func (l *Echo) Print(i ...interface{}) {
	l.logger.Print(i...)
}

// Debug string log
func (l *Echo) Debug(i ...interface{}) {
	l.logger.Debug(i...)
}

// Info string log
func (l *Echo) Info(i ...interface{}) {
	l.logger.Info(i...)
}

// Warn string log
func (l *Echo) Warn(i ...interface{}) {
	l.logger.Warn(i...)
}

// Error string log
func (l *Echo) Error(i ...interface{}) {
	l.logger.Error(i...)
}

// Fatal string log
func (l *Echo) Fatal(i ...interface{}) {
	l.logger.Fatal(i...)
}

// Panic string log
func (l *Echo) Panic(i ...interface{}) {
	l.logger.Panic(i...)
}

// Printf print JSON log
func (l *Echo) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// Debugf debug JSON log
func (l *Echo) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Infof info JSON log
func (l *Echo) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf warning JSON log
func (l *Echo) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf error JSON log
func (l *Echo) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatalf fatal JSON log
func (l *Echo) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// Panicf panic JSON log
func (l *Echo) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

// AccessLogger is a function to get a middleware to log accesses
func (l *Echo) AccessLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
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

			logger := GetLoggerFromContext(c.Request().Context())
			if logger == nil {
				logger = l.logger
			}
			logger = logger.WithCaller(false)

			// incoming log
			logger.Infow(
				fmt.Sprintf("<-- %s %s", req.Method, req.URL.Path),
				args...,
			)

			if err := next(c); err != nil {
				c.Error(err)
			}

			res = c.Response()
			stop := time.Now()
			latency := stop.Sub(start)
			latencyHuman := latency.String()
			args = append(args,
				"status", res.Status,
				"bytes_out", res.Size,
				"latency", latency.Microseconds(),
				"latency_human", latencyHuman,
			)

			// outcoming log
			logger.Infow(
				fmt.Sprintf("--> %s %d %s %s", req.Method, res.Status, req.URL.Path, latencyHuman),
				args...,
			)
			return nil
		}
	}
}

func fromMap(m map[string]any) (res []any) {
	entries := util.SortedEntries(m)
	for k, v := range entries {
		res = append(res, k)
		res = append(res, v)
	}
	return
}

func GetReqestID(w http.ResponseWriter, r *http.Request) string {
	if reqid := getHeader(r,
		"X-Request-ID",
		"X-Cloud-Trace-Context", // Google Cloud
		"X-Amzn-Trace-Id",       // AWS
		"X-ARR-LOG-ID",          // Azure
	); reqid != "" {
		return reqid
	}

	if reqid := w.Header().Get("X-Request-ID"); reqid != "" {
		return reqid
	}

	return ""
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
