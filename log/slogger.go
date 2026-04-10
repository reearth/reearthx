package log

import (
	"context"
	"log/slog"

	"go.uber.org/zap/zapcore"
)

// zapSlogHandler implements slog.Handler backed by *Logger.
type zapSlogHandler struct {
	logger *Logger
	attrs  []slog.Attr
	group  string
}

// NewSlogHandler returns a slog.Handler backed by the given *Logger.
func NewSlogHandler(l *Logger) slog.Handler {
	return &zapSlogHandler{logger: l.AddCallerSkip(2)}
}

// NewSlogLogger returns a *slog.Logger backed by the given *Logger.
func NewSlogLogger(l *Logger) *slog.Logger {
	return slog.New(NewSlogHandler(l))
}

func (h *zapSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	var zapLevel zapcore.Level
	switch {
	case level >= slog.LevelError:
		zapLevel = zapcore.ErrorLevel
	case level >= slog.LevelWarn:
		zapLevel = zapcore.WarnLevel
	case level >= slog.LevelInfo:
		zapLevel = zapcore.InfoLevel
	default:
		zapLevel = zapcore.DebugLevel
	}
	return h.logger.Level() <= zapLevel
}

func (h *zapSlogHandler) Handle(_ context.Context, r slog.Record) error {
	kvs := attrsToKVs(h.group, h.attrs)
	r.Attrs(func(a slog.Attr) bool {
		kvs = appendAttr(kvs, h.group, a)
		return true
	})

	msg := r.Message
	switch {
	case r.Level >= slog.LevelError:
		h.logger.Errorw(msg, kvs...)
	case r.Level >= slog.LevelWarn:
		h.logger.Warnw(msg, kvs...)
	case r.Level >= slog.LevelInfo:
		h.logger.Infow(msg, kvs...)
	default:
		h.logger.Debugw(msg, kvs...)
	}
	return nil
}

func (h *zapSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &zapSlogHandler{logger: h.logger, attrs: newAttrs, group: h.group}
}

func (h *zapSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	group := name
	if h.group != "" {
		group = h.group + "." + name
	}
	return &zapSlogHandler{logger: h.logger, attrs: h.attrs, group: group}
}

func attrsToKVs(group string, attrs []slog.Attr) []any {
	kvs := make([]any, 0, len(attrs)*2)
	for _, a := range attrs {
		kvs = appendAttr(kvs, group, a)
	}
	return kvs
}

func appendAttr(kvs []any, group string, a slog.Attr) []any {
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	return append(kvs, key, a.Value.Any())
}
