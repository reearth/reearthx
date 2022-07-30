package log

import (
	"go.uber.org/zap/zapcore"
)

type severity string

const (
	severityDEBUG     severity = "DEBUG"
	severityINFO      severity = "INFO"
	severityWARNING   severity = "WARNING"
	severityERROR     severity = "ERROR"
	severityCRITICAL  severity = "CRITICAL"
	severityALERT     severity = "ALERT"
	severityEmergency severity = "EMERGENCY"
)

var (
	levelsZapToGCE = map[zapcore.Level]severity{
		zapcore.DebugLevel:  severityDEBUG,
		zapcore.InfoLevel:   severityINFO,
		zapcore.WarnLevel:   severityWARNING,
		zapcore.ErrorLevel:  severityERROR,
		zapcore.PanicLevel:  severityCRITICAL,
		zapcore.DPanicLevel: severityALERT,
		zapcore.FatalLevel:  severityEmergency,
	}
)

var (
	gceEncoderConfig = zapcore.EncoderConfig{
		MessageKey:    "message",
		LevelKey:      "severity",
		TimeKey:       "time",
		NameKey:       "N",
		CallerKey:     "C",
		StacktraceKey: "S",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(string(levelsZapToGCE[l]))
		},
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
)
