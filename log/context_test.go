package log

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextLogger(t *testing.T) {
	w := &bytes.Buffer{}
	SetOutput(w)
	t.Cleanup(func() {
		SetOutput(DefaultOutput)
	})

	l := NewWithOutput(w).SetPrefix("prefix")
	ctx := AttachLoggerToContext(context.Background(), l)

	Infofc(ctx, "hoge %s", "fuga")
	Infofc(context.Background(), "hoge %s", "fuga2")
	//nolint:staticcheck // test nil context
	Infofc(nil, "hoge %s", "fuga3")

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "\tprefix\t")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga2")
	assert.NotContains(t, scanner.Text(), "\tprefix\t")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga3")
	assert.NotContains(t, scanner.Text(), "\tprefix\t")
	assert.False(t, scanner.Scan())
}
