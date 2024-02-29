package resp_decoder

import (
	"bytes"
	"io"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeSimpleString(reader *bytes.Reader) (resp_value.Value, error) {
	bytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n at the end of a simple string`,
		}
	}

	return resp_value.NewSimpleStringValue(string(bytes)), nil
}
