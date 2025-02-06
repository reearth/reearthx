package log

import "context"

type key struct{}

func AttachLoggerToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

func GetLoggerFromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return nil
	}
	if logger, ok := ctx.Value(key{}).(*Logger); ok {
		return logger
	}
	return nil
}

func GetLoggerFromContextOrDefault(ctx context.Context) *Logger {
	if logger := GetLoggerFromContext(ctx); logger != nil {
		return logger
	}
	return globalLogger
}

func UpdateContext(ctx context.Context, f func(logger *Logger) *Logger) context.Context {
	return AttachLoggerToContext(ctx, f(GetLoggerFromContextOrDefault(ctx)))
}

func WithPrefixMessage(ctx context.Context, prefix string) context.Context {
	return UpdateContext(ctx, func(logger *Logger) *Logger {
		return logger.AppendPrefixMessage(prefix)
	})
}

func Tracefc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Debugf(format, args...)
}

func Debugfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Debugf(format, args...)
}

func Infofc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Infof(format, args...)
}

func Printfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Infof(format, args...)
}

func Warnfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Warnf(format, args...)
}

func Errorfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Errorf(format, args...)
}

func Fatalfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Fatalf(format, args...)
}

func Panicfc(ctx context.Context, format string, args ...any) {
	getLoggerFromContextOrDefault(ctx).Panicf(format, args...)
}

func Tracec(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Debug(args...)
}

func Debugc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Debug(args...)
}

func Infoc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Info(args...)
}

func Printc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Info(args...)
}

func Warnc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Warn(args...)
}

func Errorc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Error(args...)
}

func Fatalc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Fatal(args...)
}

func Panicc(ctx context.Context, args ...any) {
	getLoggerFromContextOrDefault(ctx).Panic(args...)
}

func getLoggerFromContextOrDefault(ctx context.Context) *Logger {
	if logger := GetLoggerFromContext(ctx); logger != nil {
		return logger.AddCallerSkip(1)
	}
	return globalLogger
}
