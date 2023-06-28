package log

import (
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"go.uber.org/zap/zapcore"
)

type KeyValue struct {
	Key   string
	Value any
}

func (k KeyValue) interfaces() []any {
	return []any{k.Key, k.Value}
}

type Echo struct {
	logger                 *Logger
	accessLogExtraMessages func(c echo.Context) []KeyValue
}

var _ echo.Logger = (*Echo)(nil)

// NewEcho returns a logger for echo
func NewEcho() *Echo {
	return &Echo{
		logger: New(),
	}
}

func (l *Echo) SetAccessLogExtraMessages(e func(c echo.Context) []KeyValue) {
	l.accessLogExtraMessages = e
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
			if err := next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			var ex []any
			if l.accessLogExtraMessages != nil {
				ex = lo.FlatMap(l.accessLogExtraMessages(c), func(k KeyValue, _ int) []any { return k.interfaces() })
			}

			globalLogger.logger.Infow(
				"Handled request",
				append(
					[]any{
						"remote_ip", c.RealIP(),
						"host", req.Host,
						"uri", req.RequestURI,
						"method", req.Method,
						"path", req.URL.Path,
						"referer", req.Referer(),
						"user_agent", req.UserAgent(),
						"status", res.Status,
						"latency", stop.Sub(start).Microseconds(),
						"latency_human", stop.Sub(start).String(),
						"bytes_in", req.ContentLength,
						"bytes_out", res.Size,
					},
					ex...),
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
