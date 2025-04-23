package log

import (
	"io"

	"go.uber.org/zap/zapcore"
)

var (
	globalLogger = New().AddCallerSkip(1)
)

func SetLevel(l zapcore.Level) {
	globalLogger.SetLevel(l)
}

func SetOutput(w io.Writer) {
	globalLogger = NewWithOutput(w)
}

func Tracef(format string, args ...any) {
	globalLogger.Debugf(format, args...)
}

func Debugf(format string, args ...any) {
	globalLogger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	globalLogger.Infof(format, args...)
}

func Printf(format string, args ...any) {
	globalLogger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	globalLogger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	globalLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	globalLogger.Fatalf(format, args...)
}

func Panicf(format string, args ...any) {
	globalLogger.Panicf(format, args...)
}

func Trace(args ...any) {
	globalLogger.Debug(args...)
}

func Debug(args ...any) {
	globalLogger.Debug(args...)
}

func Info(args ...any) {
	globalLogger.Info(args...)
}

func Print(args ...any) {
	globalLogger.Info(args...)
}

func Warn(args ...any) {
	globalLogger.Warn(args...)
}

func Error(args ...any) {
	globalLogger.Error(args...)
}

func Fatal(args ...any) {
	globalLogger.Fatal(args...)
}

func Panic(args ...any) {
	globalLogger.Panic(args...)
}
