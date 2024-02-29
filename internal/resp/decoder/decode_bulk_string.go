package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func decodeBulkString(reader *bytes.Reader) (resp_value.Value, error) {
	offsetBeforeLength := getReaderOffset(reader)

	lengthBytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return resp_value.Value{}, IncompleteRESPError{
			Reader:  reader,
			Message: `Expected \r\n after bulk string length`,
		}
	}

	length, err := strconv.Atoi(string(lengthBytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return resp_value.Value{}, InvalidRESPError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid bulk string length: %q, expected a number", string(lengthBytes)),
		}
	}

	bytes := bytes.NewBuffer([]byte{})
	for i := 0; i < length; i++ {
		b, err := reader.ReadByte()

		if err == io.EOF {
			return resp_value.Value{}, IncompleteRESPError{
				Reader:  reader,
				Message: fmt.Sprintf("Expected %d bytes of data in bulk string, got %d", length, i),
			}
		}

		bytes.WriteByte(b)
	}

	// Read the \r\n at the end of the bulk string
	if err := readCRLF(reader, "at the end of a bulk string"); err != nil {
		return resp_value.Value{}, err
	}

	return resp_value.NewBulkStringValue(bytes.String()), nil
}
