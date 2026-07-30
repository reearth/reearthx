package log

import (
	"flag"

	"go.uber.org/zap/zapcore"
)

// Level represents a logging severity, without exposing the underlying
// zap/zapcore types to consumers of this package.
type Level int8

var _ flag.Value = (*Level)(nil)

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

// Set implements flag.Value so Level can be used directly as a flag target.
func (l *Level) Set(s string) error {
	zl := l.zap()
	if err := zl.Set(s); err != nil {
		return err
	}
	*l = Level(zl)
	return nil
}

func (l Level) zap() zapcore.Level {
	return zapcore.Level(l)
}

func levelFromZap(l zapcore.Level) Level {
	return Level(l)
}

