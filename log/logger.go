package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	consoleEncoderConfig = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "name",
		CallerKey:      "call",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	DefaultLevel  = zap.DebugLevel
	DefaultOutput = os.Stdout
)

type Logger struct {
	logger    *zap.SugaredLogger
	atom      zap.AtomicLevel
	prefix    string
	dynPrefix func() Format
	dynSuffix func() Format
}

func New() *Logger {
	return NewWithOutput(nil)
}

func NewWithOutput(w io.Writer) *Logger {
	atom := zap.NewAtomicLevelAt(DefaultLevel)
	return &Logger{
		logger: newLogger(w, atom, ""),
		atom:   atom,
	}
}

func newLogger(w io.Writer, atom zap.AtomicLevel, name string) *zap.SugaredLogger {
	if w == nil {
		w = DefaultOutput
	}

	return zap.New(
		zapcore.NewCore(
			encoder(),
			zapcore.Lock(zapcore.AddSync(w)),
			atom,
		),
	).Sugar().Named(name).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

func encoder() zapcore.Encoder {
	if isGCP() {
		return zapcore.NewJSONEncoder(gceEncoderConfig)
	} else {
		conf := consoleEncoderConfig
		if isColorDisabled() {
			conf.EncodeLevel = zapcore.CapitalLevelEncoder
		}
		return zapcore.NewConsoleEncoder(conf)
	}
}

func (l *Logger) AppendPrefixMessage(prefix string) *Logger {
	return l.AppendDynamicPrefix(func() Format {
		return Format{Format: prefix}
	})
}

func (l *Logger) AppendSuffixMessage(suffix string) *Logger {
	return l.AppendDynamicSuffix(func() Format {
		return Format{Format: suffix}
	})
}

func (l *Logger) AppendDynamicPrefix(prefix func() Format) *Logger {
	if l.dynPrefix == nil {
		return l.SetDynamicPrefix(prefix)
	}

	return l.SetDynamicPrefix(func() Format {
		if l.dynPrefix == nil {
			return prefix()
		}
		return l.dynPrefix().Append(prefix())
	})
}

func (l *Logger) AppendDynamicSuffix(suffix func() Format) *Logger {
	if l.dynSuffix == nil {
		return l.SetDynamicSuffix(suffix)
	}

	return l.SetDynamicSuffix(func() Format {
		if l.dynSuffix == nil {
			return suffix()
		}
		return l.dynSuffix().Append(suffix())
	})
}

func (l *Logger) SetDynamicPrefix(prefix func() Format) *Logger {
	return &Logger{
		logger:    l.logger,
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: prefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) SetDynamicSuffix(suffix func() Format) *Logger {
	return &Logger{
		logger:    l.logger,
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: suffix,
	}
}

func (l *Logger) SetOutput(w io.Writer) *Logger {
	return &Logger{
		logger:    newLogger(w, l.atom, l.prefix),
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) Level() zapcore.Level {
	return l.atom.Level()
}

func (l *Logger) SetLevel(lv zapcore.Level) {
	l.atom.SetLevel(lv)
}

func (l *Logger) Prefix() string {
	return l.prefix
}

func (l *Logger) SetPrefix(prefix string) *Logger {
	if prefix == "" {
		return l
	}
	return &Logger{
		logger:    l.logger.Named(prefix),
		atom:      l.atom,
		prefix:    prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) ClearPrefix() *Logger {
	return &Logger{
		logger: l.logger.Named(""),
		atom:   l.atom,
		prefix: "",
	}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger:    l.logger.With(args...),
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) WithCaller(enabled bool) *Logger {
	return &Logger{
		logger:    l.logger.WithOptions(zap.WithCaller(enabled)),
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) AddCallerSkip(skip int) *Logger {
	return &Logger{
		logger:    l.logger.WithOptions(zap.AddCallerSkip(skip)),
		atom:      l.atom,
		prefix:    l.prefix,
		dynPrefix: l.dynPrefix,
		dynSuffix: l.dynSuffix,
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Debugf(f.Format, f.Args...)
}

func (l *Logger) Infof(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Infof(f.Format, f.Args...)
}

func (l *Logger) Printf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Infof(f.Format, f.Args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Warnf(f.Format, f.Args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Errorf(f.Format, f.Args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Fatalf(f.Format, f.Args...)
}

func (l *Logger) Panicf(format string, args ...any) {
	f := l.format(format, args...)
	l.logger.Panicf(f.Format, f.Args...)
}

func (l *Logger) Debugw(msg string, keyAndValues ...any) {
	l.logger.Debugw(l.msg(msg), keyAndValues...)
}

func (l *Logger) Infow(msg string, keyAndValues ...any) {
	l.logger.Infow(l.msg(msg), keyAndValues...)
}

func (l *Logger) Printw(msg string, keyAndValues ...any) {
	l.logger.Infow(l.msg(msg), keyAndValues...)
}

func (l *Logger) Warnw(msg string, keyAndValues ...any) {
	l.logger.Warnw(l.msg(msg), keyAndValues...)
}

func (l *Logger) Errorw(msg string, keyAndValues ...any) {
	l.logger.Errorw(l.msg(msg), keyAndValues...)
}

func (l *Logger) Fatalw(msg string, keyAndValues ...any) {
	l.logger.Fatalw(l.msg(msg), keyAndValues...)
}

func (l *Logger) Panicw(msg string, keyAndValues ...any) {
	l.logger.Panicw(l.msg(msg), keyAndValues...)
}

func (l *Logger) Debug(args ...any) {
	l.logger.Debug(l.args(args)...)
}

func (l *Logger) Info(args ...any) {
	l.logger.Info(l.args(args)...)
}

func (l *Logger) Print(args ...any) {
	l.logger.Info(l.args(args)...)
}

func (l *Logger) Warn(args ...any) {
	l.logger.Warn(l.args(args)...)
}

func (l *Logger) Error(args ...any) {
	l.logger.Error(l.args(args)...)
}

func (l *Logger) Fatal(args ...any) {
	l.logger.Fatal(l.args(args)...)
}

func (l *Logger) Panic(args ...any) {
	l.logger.Panic(l.args(args)...)
}

func (l *Logger) format(format string, args ...any) Format {
	f := Format{
		Format: format,
		Args:   args,
	}

	if l.dynPrefix != nil {
		f = f.Prepend(l.dynPrefix())
	}

	if l.dynSuffix != nil {
		f = f.Append(l.dynSuffix())
	}

	return f
}

func (l *Logger) args(args ...any) []any {
	p, s := "", ""
	if l.dynPrefix != nil {
		p = l.dynPrefix().String()
	}

	if l.dynSuffix != nil {
		s = l.dynSuffix().String()
	}

	if p != "" && s != "" {
		return append([]any{p}, append(args, s)...)
	} else if p != "" {
		return append([]any{p}, args...)
	} else if s != "" {
		return append(args, s)
	}
	return args
}

func (l *Logger) msg(msg string) string {
	p, s := "", ""
	if l.dynPrefix != nil {
		p = l.dynPrefix().String()
	}
	if l.dynSuffix != nil {
		s = l.dynSuffix().String()
	}
	return p + msg + s
}

func isColorDisabled() bool {
	return os.Getenv("NO_COLOR") != ""
}
