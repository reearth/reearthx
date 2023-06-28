package log

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewEcho()
	l.SetOutput(w)
	l.SetPrefix("test")
	l.Infof("hoge %s", "fuga")

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "test\thoge fuga")
	assert.False(t, scanner.Scan())
}
