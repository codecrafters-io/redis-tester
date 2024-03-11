package resp_decoder

import (
	"bytes"
	"io"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeError(reader *bytes.Reader) (resp_value.Value, error) {
	bytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n at the end of a simple error`,
		}
	}

	return resp_value.NewErrorValue(string(bytes)), nil
}
