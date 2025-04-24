package log

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)
	l.Infow("msg", "aaaa", 123)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "fuga 1")
	assert.True(t, scanner.Scan())
	assert.Contains(t, stripansi.Strip(scanner.Text()), "msg")
	assert.Contains(t, stripansi.Strip(scanner.Text()), "aaaa")
	assert.Contains(t, stripansi.Strip(scanner.Text()), "123")
	assert.False(t, scanner.Scan())
}

func TestLogger_SetPrefix(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w).SetPrefix("test")
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)
	l.Infow("fuga", "abcd", 123)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Regexp(t, `\ttest\t.+?\thoge fuga$`, scanner.Text()) // .+? is a caller
	assert.True(t, scanner.Scan())
	assert.Regexp(t, `\ttest\t.+?\t\[fuga 1\]$`, scanner.Text())
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "test")
	assert.Contains(t, scanner.Text(), "abcd")
	assert.Contains(t, scanner.Text(), "123")
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
			Args:   []any{"suffix"},
		}
	})
	l.Infof("hoge %s", "fuga")
	l.Info("fuga", 1)

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "[test] hoge fuga <suffix>")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "[test] [fuga 1] <suffix>")
	assert.False(t, scanner.Scan())
}

func TestLogger_AppendDynamicPrefixS(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w).AppendDynamicPrefix(func() Format {
		return Format{
			Format: "[%s] ",
			Args:   []any{"prefix"},
		}
	}).AppendPrefixMessage("<prefix2> ").AppendDynamicSuffix(func() Format {
		return Format{
			Format: " [%s]",
			Args:   []any{"suffix"},
		}
	}).AppendSuffixMessage(" <suffix2>")
	l.Infof("hoge %s", "fuga")

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "[prefix] <prefix2> hoge fuga [suffix] <suffix2>")
	assert.False(t, scanner.Scan())
}
