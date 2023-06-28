package log

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "fuga 1")
	assert.False(t, scanner.Scan())
}

func TestLogger_SetPrefix(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w).SetPrefix("test")
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "test\thoge fuga")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "test\t[fuga 1]")
	assert.False(t, scanner.Scan())
}

func TestLogger_DynamicPrefixSuffix(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w).SetDynamicPrefix(func() Format {
		return Format{
			Format: "[%s] ",
			Args:   []any{"test"},
		}
	}).SetDynamicSuffix(func() Format {
		return Format{
			Format: " <%s>",
			Args:   []any{"prefix"},
		}
	})
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "[test] hoge fuga <prefix>")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "[test] [fuga 1] <prefix>")
	assert.False(t, scanner.Scan())
}
