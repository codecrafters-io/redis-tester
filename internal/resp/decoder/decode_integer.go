package resp_decoder

import (
	"bytes"
	"io"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeInteger(reader *bytes.Reader) (resp_value.Value, error) {
	bytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n at the end of an integer`,
		}
	}

	integer, err := strconv.Atoi(string(bytes))
	if err != nil {
		return resp_value.Value{}, err
	}
	return resp_value.NewIntegerValue(integer), nil
}
