package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeBulkStringOrNil(reader *bytes.Reader) (resp_value.Value, error) {
	offsetBeforeLength := getReaderOffset(reader)

	lengthBytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n after bulk string length`,
		}
	}

	length, err := strconv.Atoi(string(lengthBytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid bulk string length: %q, expected a number", string(lengthBytes)),
		}
	}

	if length == -1 {
		return resp_value.NewNilValue(), nil
	}

	if length < 0 {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid bulk string length: %d, expected a positive integer", length),
		}
	}

	bytes := bytes.NewBuffer([]byte{})
	for i := 0; i < length; i++ {
		b, err := reader.ReadByte()

		if err == io.EOF {
			return resp_value.Value{}, IncompleteInputError{
				Reader:  reader,
				Message: fmt.Sprintf("Expected %d bytes of data in bulk string, got %d", length, i),
			}
		}

		bytes.WriteByte(b)
	}

	// Read the \r\n at the end of the bulk string
	errorMessage := fmt.Sprintf(`Expected \r\n after %d bytes of data in bulk string`, length)
	if err := readCRLF(reader, errorMessage); err != nil {
		return resp_value.Value{}, err
	}

	return resp_value.NewBulkStringValue(bytes.String()), nil
}
