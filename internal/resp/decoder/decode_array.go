package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeArray(reader *bytes.Reader) (resp_value.Value, error) {
	offsetBeforeLength := getReaderOffset(reader)

	lengthBytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n after array length`,
		}
	}

	length, err := strconv.Atoi(string(lengthBytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid array length: %q, expected a number", string(lengthBytes)),
		}
	}

	if length == -1 {
		return resp_value.NewNilArrayValue(), nil
	}

	if length < -1 {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid array length: %d, expected 0 or a positive integer", length),
		}
	}

	values := make([]resp_value.Value, length)
	for i := 0; i < length; i++ {
		value, err := doDecodeValue(reader)
		if err != nil {
			return resp_value.Value{}, err
		}

		values[i] = value
	}

	return resp_value.NewArrayValue(values), nil
}
