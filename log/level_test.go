package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{name: "debug", level: LevelDebug, want: "debug"},
		{name: "info", level: LevelInfo, want: "info"},
		{name: "warn", level: LevelWarn, want: "warn"},
		{name: "error", level: LevelError, want: "error"},
		{name: "dpanic", level: LevelDPanic, want: "dpanic"},
		{name: "panic", level: LevelPanic, want: "panic"},
		{name: "fatal", level: LevelFatal, want: "fatal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestLevel_zap(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  zapcore.Level
	}{
		{name: "debug", level: LevelDebug, want: zapcore.DebugLevel},
		{name: "info", level: LevelInfo, want: zapcore.InfoLevel},
		{name: "warn", level: LevelWarn, want: zapcore.WarnLevel},
		{name: "error", level: LevelError, want: zapcore.ErrorLevel},
		{name: "dpanic", level: LevelDPanic, want: zapcore.DPanicLevel},
		{name: "panic", level: LevelPanic, want: zapcore.PanicLevel},
		{name: "fatal", level: LevelFatal, want: zapcore.FatalLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.level.zap())
		})
	}
}

func TestLevelFromZap(t *testing.T) {
	tests := []struct {
		name string
		in   zapcore.Level
		want Level
	}{
		{name: "debug", in: zapcore.DebugLevel, want: LevelDebug},
		{name: "info", in: zapcore.InfoLevel, want: LevelInfo},
		{name: "warn", in: zapcore.WarnLevel, want: LevelWarn},
		{name: "error", in: zapcore.ErrorLevel, want: LevelError},
		{name: "dpanic", in: zapcore.DPanicLevel, want: LevelDPanic},
		{name: "panic", in: zapcore.PanicLevel, want: LevelPanic},
		{name: "fatal", in: zapcore.FatalLevel, want: LevelFatal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, levelFromZap(tt.in))
		})
	}
}
