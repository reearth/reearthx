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

	l := NewWithOutput(w).SetPrefix("test")
	ctx := AttachLoggerToContext(context.Background(), l)
	Infofc(ctx, "hoge %s", "fuga")
	Infofc(context.Background(), "hoge %s", "fuga2")
	//nolint:staticcheck // test context.TODO() instead of nil context
	Infofc(context.TODO(), "hoge %s", "fuga3")

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "test\thoge fuga")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga2")
	assert.NotContains(t, scanner.Text(), "test")
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge fuga3")
	assert.NotContains(t, scanner.Text(), "test")
	assert.False(t, scanner.Scan())
}
