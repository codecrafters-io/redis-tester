package resp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeSimpleStringSuccess(t *testing.T) {
	value, nBytes, err := Decode([]byte("+OK\r\n"))
	assert.Nil(t, err)
	assert.Equal(t, 5, nBytes)
	assert.Equal(t, "OK", string(value.data))
	assert.Equal(t, SIMPLE_STRING, value.Type)
}

func TestDecodeWithExtraDataSuccess(t *testing.T) {
	value, nBytes, err := Decode([]byte("+OK\r\nextra"))
	assert.Nil(t, err)
	assert.Equal(t, SIMPLE_STRING, value.Type)
	assert.Equal(t, "OK", string(value.data))
	assert.Equal(t, 5, nBytes)
}

func TestDecodeIncompleteSimpleStringFailure(t *testing.T) {
	_, _, err := Decode([]byte("+OK"))
	assert.NotNil(t, err)
	incompleteRespErr, ok := err.(IncompleteRESPError)
	assert.True(t, ok)

	assert.Equal(t, strings.TrimSpace(`
Received: "+OK"
              ^
Expected \r\n at the end of a simple string.
	`), incompleteRespErr.Error())
}

func TestDecodeInvalidSimpleStringFailure(t *testing.T) {
	_, _, err := Decode([]byte("OK\r\n"))
	assert.NotNil(t, err)
	invalidRespErr, ok := err.(InvalidRESPError)
	assert.True(t, ok)

	assert.Equal(t, strings.TrimSpace(`
Received: "OK\r\n"
           ^
"O" is not a valid start of a new RESP value (expected +, -, :, $, or *)
	`), invalidRespErr.Error())
}
