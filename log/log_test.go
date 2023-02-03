package log

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	w := &bytes.Buffer{}
	SetOutput(w)
	t.Cleanup(func() {
		SetOutput(DefaultOutput)
	})

	Debug("hoge")

	scanner := bufio.NewScanner(w)
	assert.True(t, scanner.Scan())
	assert.Contains(t, scanner.Text(), "hoge")
	assert.False(t, scanner.Scan())
}
