package resp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeSuccess(t *testing.T) {
	value, err := Decode([]byte("+OK\r\n"))
	assert.Nil(t, err)
	assert.Equal(t, SIMPLE_STRING, value.Type)
	assert.Equal(t, "OK", string(value.data))
}

func TestDecodeFailure(t *testing.T) {
	_, err := Decode([]byte("OK\r\n"))
	assert.NotNil(t, err)
	invalidRespErr, ok := err.(InvalidRESPError)
	assert.True(t, ok)

	assert.Equal(t, strings.TrimSpace(`
Received: "OK\r\n"
"O" is not a valid start of a new RESP value (expected +, -, :, $, or *)
	`), invalidRespErr.DetailedError())
}
