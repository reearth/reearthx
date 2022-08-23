package log

import (
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/reearth/reearthx/util"
	"go.uber.org/zap/zapcore"
)

type Echo struct{}

var _ echo.Logger = (*Echo)(nil)

// NewEcho returns a logger for echo
func NewEcho() *Echo {
	return &Echo{}
}

// Level returns logger level
func (l *Echo) Level() log.Lvl {
	switch atom.Level() {
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
func (l *Echo) SetPrefix(s string) {}

// Prefix It's controlled by Logger
func (l *Echo) Prefix() string {
	return ""
}

// SetLevel set level to logger from given log.Lvl
func (l *Echo) SetLevel(lvl log.Lvl) {
	switch lvl {
	case log.DEBUG:
		SetLevel(zapcore.DebugLevel)
	case log.INFO:
		SetLevel(zapcore.InfoLevel)
	case log.WARN:
		SetLevel(zapcore.WarnLevel)
	case log.ERROR:
		SetLevel(zapcore.ErrorLevel)
	default:
		l.Panic("Invalid level")
	}
}

// Output logger output func
func (l *Echo) Output() io.Writer {
	return writer
}

// SetOutput change output, default os.Stdout
func (l *Echo) SetOutput(w io.Writer) {
	SetOutput(w)
}

// Printj print JSON log
func (l *Echo) Printj(j log.JSON) {
	Print(fromMap(j))
}

// Debugj debug JSON log
func (l *Echo) Debugj(j log.JSON) {
	Debug(fromMap(j))
}

// Infoj info JSON log
func (l *Echo) Infoj(j log.JSON) {
	Info(fromMap(j))
}

// Warnj warning JSON log
func (l *Echo) Warnj(j log.JSON) {
	Warn(fromMap(j))
}

// Errorj error JSON log
func (l *Echo) Errorj(j log.JSON) {
	Error(fromMap(j))
}

// Fatalj fatal JSON log
func (l *Echo) Fatalj(j log.JSON) {
	Fatal()
}

// Panicj panic JSON log
func (l *Echo) Panicj(j log.JSON) {
	Panic()
}

// Print string log
func (l *Echo) Print(i ...interface{}) {
	Print(i...)
}

// Debug string log
func (l *Echo) Debug(i ...interface{}) {
	Debug(i...)
}

// Info string log
func (l *Echo) Info(i ...interface{}) {
	Info(i...)
}

// Warn string log
func (l *Echo) Warn(i ...interface{}) {
	Warn(i...)
}

// Error string log
func (l *Echo) Error(i ...interface{}) {
	Error(i...)
}

// Fatal string log
func (l *Echo) Fatal(i ...interface{}) {
	Fatal(i...)
}

// Panic string log
func (l *Echo) Panic(i ...interface{}) {
	Panic(i...)
}

// Printf print JSON log
func (l *Echo) Printf(format string, args ...interface{}) {
	Printf(format, args...)
}

// Debugf debug JSON log
func (l *Echo) Debugf(format string, args ...interface{}) {
	Debugf(format, args...)
}

// Infof info JSON log
func (l *Echo) Infof(format string, args ...interface{}) {
	Infof(format, args...)
}

// Warnf warning JSON log
func (l *Echo) Warnf(format string, args ...interface{}) {
	Warnf(format, args...)
}

// Errorf error JSON log
func (l *Echo) Errorf(format string, args ...interface{}) {
	Errorf(format, args...)
}

// Fatalf fatal JSON log
func (l *Echo) Fatalf(format string, args ...interface{}) {
	Fatalf(format, args...)
}

// Panicf panic JSON log
func (l *Echo) Panicf(format string, args ...interface{}) {
	Panicf(format, args...)
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

			logger.Infow(
				"Handled request",
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
