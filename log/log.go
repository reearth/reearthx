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

	atom          = zap.NewAtomicLevel()
	sugaredLogger *zap.SugaredLogger
	logger        *zap.Logger
	writer        = os.Stdout
)

func init() {
	logger = zap.New(
		zapcore.NewCore(
			enc(),
			zapcore.Lock(zapcore.AddSync(writer)),
			atom,
		),
	)

	sugaredLogger = logger.Sugar()
}

func enc() zapcore.Encoder {
	gcp, _ := os.LookupEnv("GOOGLE_CLOUD_PROJECT")

	if gcp == "" {
		return zapcore.NewJSONEncoder(gceEncoderConfig)
	} else {
		return zapcore.NewConsoleEncoder(consoleEncoderConfig)
	}
}

func SetLevel(l zapcore.Level) {
	atom.SetLevel(l)
}

func SetOutput(w io.Writer) {
	logger = zap.New(
		zapcore.NewCore(
			enc(),
			zapcore.Lock(zapcore.AddSync(writer)),
			atom,
		),
	)

	sugaredLogger = logger.Sugar()
}

func Tracef(format string, args ...interface{}) {
	sugaredLogger.Debugf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	sugaredLogger.Debugf(format, args)
}

func Infof(format string, args ...interface{}) {
	sugaredLogger.Infof(format, args)
}

func Printf(format string, args ...interface{}) {
	sugaredLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugaredLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugaredLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	sugaredLogger.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	sugaredLogger.Panicf(format, args...)
}

func Trace(args ...interface{}) {
	sugaredLogger.Debug(args...)
}

func Debug(args ...interface{}) {
	sugaredLogger.Debug(args...)
}

func Info(args ...interface{}) {
	sugaredLogger.Info(args...)
}

func Print(args ...interface{}) {
	sugaredLogger.Info(args...)
}

func Warn(args ...interface{}) {
	sugaredLogger.Warn(args...)
}

func Error(args ...interface{}) {
	sugaredLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	sugaredLogger.Fatal(args...)
}

func Panic(args ...interface{}) {
	sugaredLogger.Panic(args...)
}

func Traceln(args ...interface{}) {
	sugaredLogger.Debug(args...)
}

func Debugln(args ...interface{}) {
	sugaredLogger.Debug(args...)
}

func Infoln(args ...interface{}) {
	sugaredLogger.Info(args...)
}

func Println(args ...interface{}) {
	sugaredLogger.Info(args...)
}

func Warnln(args ...interface{}) {
	sugaredLogger.Warn(args...)
}

func Errorln(args ...interface{}) {
	sugaredLogger.Error(args...)
}

func Fatalln(args ...interface{}) {
	sugaredLogger.Fatal(args...)
}

func Panicln(args ...interface{}) {
	sugaredLogger.Panic(args...)
}
