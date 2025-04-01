package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceIDFromTraceparent(t *testing.T) {
	assert.Equal(t, "1234567890abcdef1234567890abcdef", TraceIDFromTraceparent("00-1234567890abcdef1234567890abcdef-1234567890abcdef-01"))
	assert.Equal(t, "", TraceIDFromTraceparent("832E80E9-B9E9-4A91-8EA2-2E2513F55D86"))
}

func TestTraceIDFromXCloudTraceContext(t *testing.T) {
	assert.Equal(t, "1234567890abcdef1234567890abcdef", TraceIDFromXCloudTraceContext("1234567890abcdef1234567890abcdef/1234567890abcdef;o=1"))
	assert.Equal(t, "", TraceIDFromXCloudTraceContext("00-1234567890abcdef1234567890abcdef-1234567890abcdef-01"))
}
