package log

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSlogLogger(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l)

	sl.Info("hello slog", "key", "value")

	out := w.String()
	assert.Contains(t, out, "hello slog")
	assert.Contains(t, out, "key")
	assert.Contains(t, out, "value")
}

func TestSlogHandler_Levels(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l)

	sl.Debug("debug msg")
	sl.Info("info msg")
	sl.Warn("warn msg")
	sl.Error("error msg")

	out := w.String()
	assert.Contains(t, out, "debug msg")
	assert.Contains(t, out, "info msg")
	assert.Contains(t, out, "warn msg")
	assert.Contains(t, out, "error msg")
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l).With("persistent", "attr")

	sl.Info("msg with attr")

	out := w.String()
	assert.Contains(t, out, "persistent")
	assert.Contains(t, out, "attr")
}

func TestSlogHandler_WithGroup(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l).WithGroup("grp")

	sl.Info("grouped", slog.String("key", "val"))

	out := w.String()
	assert.Contains(t, out, "grp.key")
	assert.Contains(t, out, "val")
}
