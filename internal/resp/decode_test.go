package resp

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

type DecodeErrorTestCase struct {
	Input string `yaml:"input"`
	Error string `yaml:"error"`
}

func TestDecodeErrors(t *testing.T) {
	testCases := []DecodeErrorTestCase{}

	yamlContents, err := os.ReadFile("decode_error_tests.yml")
	assert.Nil(t, err)

	err = yaml.Unmarshal(yamlContents, &testCases)
	assert.Nil(t, err)

	for _, testCase := range testCases {
		_, _, err := Decode([]byte(testCase.Input))
		assert.NotNil(t, err)
		assert.Equal(t, strings.TrimSpace(testCase.Error), err.Error())
	}
}
