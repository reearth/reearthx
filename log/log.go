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

	atom   = zap.NewAtomicLevel()
	logger *zap.SugaredLogger
	writer = os.Stdout
)

func init() {
	l := zap.New(
		zapcore.NewCore(
			enc(),
			zapcore.Lock(zapcore.AddSync(writer)),
			atom,
		),
	)

	logger = l.Sugar()
}

func enc() zapcore.Encoder {
	gcp, _ := os.LookupEnv("GOOGLE_CLOUD_PROJECT")

	if gcp == "" {
		return zapcore.NewConsoleEncoder(consoleEncoderConfig)
	} else {
		return zapcore.NewJSONEncoder(gceEncoderConfig)
	}
}

func SetLevel(l zapcore.Level) {
	atom.SetLevel(l)
}

func SetOutput(w io.Writer) {
	l := zap.New(
		zapcore.NewCore(
			enc(),
			zapcore.Lock(zapcore.AddSync(writer)),
			atom,
		),
	)

	logger = l.Sugar()
}

func Tracef(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args)
}

func Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func Trace(args ...interface{}) {
	logger.Debug(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Print(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Traceln(args ...interface{}) {
	logger.Debug(args...)
}

func Debugln(args ...interface{}) {
	logger.Debug(args...)
}

func Infoln(args ...interface{}) {
	logger.Info(args...)
}

func Println(args ...interface{}) {
	logger.Info(args...)
}

func Warnln(args ...interface{}) {
	logger.Warn(args...)
}

func Errorln(args ...interface{}) {
	logger.Error(args...)
}

func Fatalln(args ...interface{}) {
	logger.Fatal(args...)
}

func Panicln(args ...interface{}) {
	logger.Panic(args...)
}
