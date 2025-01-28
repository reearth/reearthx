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
		return logger.AddCallerSkip(1)
	}
	return globalLogger.AddCallerSkip(1)
}

func UpdateContext(ctx context.Context, f func(logger *Logger) *Logger) context.Context {
	return AttachLoggerToContext(ctx, f(GetLoggerFromContextOrDefault(ctx)))
}

func Tracefc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Debugf(format, args...)
}

func Debugfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Debugf(format, args...)
}

func Infofc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Infof(format, args...)
}

func Printfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Infof(format, args...)
}

func Warnfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Warnf(format, args...)
}

func Errorfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Errorf(format, args...)
}

func Fatalfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Fatalf(format, args...)
}

func Panicfc(ctx context.Context, format string, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Panicf(format, args...)
}

func Tracec(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Debug(args...)
}

func Debugc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Debug(args...)
}

func Infoc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Info(args...)
}

func Printc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Info(args...)
}

func Warnc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Warn(args...)
}

func Errorc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Error(args...)
}

func Fatalc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Fatal(args...)
}

func Panicc(ctx context.Context, args ...any) {
	GetLoggerFromContextOrDefault(ctx).Panic(args...)
}
