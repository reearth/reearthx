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
	atom          = zap.NewAtomicLevelAt(DefaultLevel)
	logger        = new(DefaultOutput)
)

func SetLevel(l zapcore.Level) {
	atom.SetLevel(l)
}

func SetOutput(w io.Writer) {
	logger = new(w)
}

func new(w io.Writer) *zap.SugaredLogger {
	return zap.New(
		zapcore.NewCore(
			enc(),
			zapcore.Lock(zapcore.AddSync(w)),
			atom,
		),
	).Sugar()
}

func enc() zapcore.Encoder {
	if isGCP() {
		return zapcore.NewJSONEncoder(gceEncoderConfig)
	} else {
		return zapcore.NewConsoleEncoder(consoleEncoderConfig)
	}
}

func Tracef(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Printf(format string, args ...any) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

func Trace(args ...any) {
	logger.Debug(args...)
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Print(args ...any) {
	logger.Info(args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Fatal(args ...any) {
	logger.Fatal(args...)
}

func Panic(args ...any) {
	logger.Panic(args...)
}

func Traceln(args ...any) {
	logger.Debug(args...)
}

func Debugln(args ...any) {
	logger.Debug(args...)
}

func Infoln(args ...any) {
	logger.Info(args...)
}

func Println(args ...any) {
	logger.Info(args...)
}

func Warnln(args ...any) {
	logger.Warn(args...)
}

func Errorln(args ...any) {
	logger.Error(args...)
}

func Fatalln(args ...any) {
	logger.Fatal(args...)
}

func Panicln(args ...any) {
	logger.Panic(args...)
}
