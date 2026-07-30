package log

import "go.uber.org/zap/zapcore"

// Level represents a logging severity, without exposing the underlying
// zap/zapcore types to consumers of this package.
type Level int8

const (
	LevelDebug  Level = Level(zapcore.DebugLevel)
	LevelInfo   Level = Level(zapcore.InfoLevel)
	LevelWarn   Level = Level(zapcore.WarnLevel)
	LevelError  Level = Level(zapcore.ErrorLevel)
	LevelDPanic Level = Level(zapcore.DPanicLevel)
	LevelPanic  Level = Level(zapcore.PanicLevel)
	LevelFatal  Level = Level(zapcore.FatalLevel)
)

func (l Level) String() string {
	return l.zap().String()
}

func (l Level) zap() zapcore.Level {
	return zapcore.Level(l)
}

func levelFromZap(l zapcore.Level) Level {
	return Level(l)
}
