package response

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	// Test: HTTP OK
	buf := &bytes.Buffer{}
	err := WriteStatusLine(buf, 200)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 200 OK\n", buf.String())
}

func TestDefaultHeaders(t *testing.T) {
	// Test response is as expected
	h := GetDefaultHeaders(10)
	assert.Equal(t, len(h), 3)
	cl, _ := h["Content-Length"]
	assert.Equal(t, "10", cl)
}
