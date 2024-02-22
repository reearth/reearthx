package log

import (
	"os"

	"go.uber.org/zap/zapcore"
)

var GCP = false
var GCPEnv = []string{
	"CLOUD_RUN_JOB",
	"K_SERVICE",
	"GOOGLE_CLOUD_PROJECT",
	"GCP",
}

type severity string

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
const (
	severityDEBUG     severity = "DEBUG"
	severityINFO      severity = "INFO"
	severityWARNING   severity = "WARNING"
	severityERROR     severity = "ERROR"
	severityCRITICAL  severity = "CRITICAL"
	severityALERT     severity = "ALERT"
	severityEMERGENCY severity = "EMERGENCY"
)

var levelsZapToGCE = map[zapcore.Level]severity{
	zapcore.DebugLevel:  severityDEBUG,
	zapcore.InfoLevel:   severityINFO,
	zapcore.WarnLevel:   severityWARNING,
	zapcore.ErrorLevel:  severityERROR,
	zapcore.PanicLevel:  severityCRITICAL,
	zapcore.DPanicLevel: severityALERT,
	zapcore.FatalLevel:  severityEMERGENCY,
}

var gceEncoderConfig = zapcore.EncoderConfig{
	MessageKey:    "message",
	LevelKey:      "severity",
	TimeKey:       "time",
	NameKey:       "name",
	CallerKey:     "call",
	StacktraceKey: "stack",
	LineEnding:    zapcore.DefaultLineEnding,
	EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(string(levelsZapToGCE[l]))
	},
	// https://github.com/GoogleCloudPlatform/fluent-plugin-google-cloud/issues/220#issuecomment-382505054
	EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeName:     zapcore.FullNameEncoder,
}

func isGCP() bool {
	if GCP {
		return true
	}

	for _, key := range GCPEnv {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return true
		}
	}

	return false
}
