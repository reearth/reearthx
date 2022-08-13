package log

import (
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap/zapcore"
)

type Logger struct{}

var _ echo.Logger = (*Logger)(nil)

// GetEchoLogger returns Logger
func GetEchoLogger() *Logger {
	return &Logger{}
}

// Level returns logger level
func (l *Logger) Level() log.Lvl {
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
func (l *Logger) SetHeader(_ string) {}

// SetPrefix It's controlled by Logger
func (l *Logger) SetPrefix(s string) {}

// Prefix It's controlled by Logger
func (l *Logger) Prefix() string {
	return ""
}

// SetLevel set level to logger from given log.Lvl
func (l *Logger) SetLevel(lvl log.Lvl) {
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
func (l *Logger) Output() io.Writer {
	return writer
}

// SetOutput change output, default os.Stdout
func (l *Logger) SetOutput(w io.Writer) {
	SetOutput(w)
}

// Printj print JSON log
func (l *Logger) Printj(j log.JSON) {
	Print(fromMap(j))
}

// Debugj debug JSON log
func (l *Logger) Debugj(j log.JSON) {
	Debug(fromMap(j))
}

// Infoj info JSON log
func (l *Logger) Infoj(j log.JSON) {
	Info(fromMap(j))
}

// Warnj warning JSON log
func (l *Logger) Warnj(j log.JSON) {
	Warn(fromMap(j))
}

// Errorj error JSON log
func (l *Logger) Errorj(j log.JSON) {
	Error(fromMap(j))
}

// Fatalj fatal JSON log
func (l *Logger) Fatalj(j log.JSON) {
	Fatal()
}

// Panicj panic JSON log
func (l *Logger) Panicj(j log.JSON) {
	Panic()
}

// Print string log
func (l *Logger) Print(i ...interface{}) {
	Print(i...)
}

// Debug string log
func (l *Logger) Debug(i ...interface{}) {
	Debug(i...)
}

// Info string log
func (l *Logger) Info(i ...interface{}) {
	Info(i...)
}

// Warn string log
func (l *Logger) Warn(i ...interface{}) {
	Warn(i...)
}

// Error string log
func (l *Logger) Error(i ...interface{}) {
	Error(i...)
}

// Fatal string log
func (l *Logger) Fatal(i ...interface{}) {
	Fatal(i...)
}

// Panic string log
func (l *Logger) Panic(i ...interface{}) {
	Panic(i...)
}

// Printf print JSON log
func (l *Logger) Printf(format string, args ...interface{}) {
	Printf(format, args...)
}

// Debugf debug JSON log
func (l *Logger) Debugf(format string, args ...interface{}) {
	Debugf(format, args...)
}

// Infof info JSON log
func (l *Logger) Infof(format string, args ...interface{}) {
	Infof(format, args...)
}

// Warnf warning JSON log
func (l *Logger) Warnf(format string, args ...interface{}) {
	Warnf(format, args...)
}

// Errorf error JSON log
func (l *Logger) Errorf(format string, args ...interface{}) {
	Errorf(format, args...)
}

// Fatalf fatal JSON log
func (l *Logger) Fatalf(format string, args ...interface{}) {
	Fatalf(format, args...)
}

// Panicf panic JSON log
func (l *Logger) Panicf(format string, args ...interface{}) {
	Panicf(format, args...)
}

// Hook is a function to process middleware.
func (l *Logger) Hook() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			fs := map[string]interface{}{
				"remote_ip":     c.RealIP(),
				"host":          req.Host,
				"uri":           req.RequestURI,
				"method":        req.Method,
				"path":          req.URL.Path,
				"referer":       req.Referer(),
				"user_agent":    req.UserAgent(),
				"status":        res.Status,
				"latency":       stop.Sub(start).Microseconds(),
				"latency_human": stop.Sub(start).String(),
				"bytes_in":      req.ContentLength,
				"bytes_out":     res.Size,
			}

			logger.Infow("Handled request", fromMap(fs)...)
			return nil
		}
	}
}

func fromMap(fs map[string]interface{}) []interface{} {
	fields := make([]interface{}, 0, len(fs))
	for k, v := range fs {
		fields = append(fields, k)
		fields = append(fields, v)
	}

	return fields
}
