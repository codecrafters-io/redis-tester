package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeInteger(reader *bytes.Reader) (resp_value.Value, error) {
	offsetBeforeInteger := getReaderOffset(reader)

	bytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n at the end of an integer`,
		}
	}

	integer, err := strconv.Atoi(string(bytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeInteger), io.SeekStart)

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid integer: %q, expected a number", string(bytes)),
		}
	}

	return resp_value.NewIntegerValue(integer), nil
}
