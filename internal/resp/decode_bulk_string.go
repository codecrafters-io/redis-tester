package resp

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

func decodeBulkString(reader *bytes.Reader) (Value, error) {
	offsetBeforeLength := getReaderOffset(reader)

	lengthBytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return Value{}, IncompleteRESPError{
			Reader:  reader,
			Message: `Expected \r\n after bulk string length`,
		}
	}

	length, err := strconv.Atoi(string(lengthBytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return Value{}, InvalidRESPError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid bulk string length: %q, expected a number", string(lengthBytes)),
		}
	}

	bytes := bytes.NewBuffer([]byte{})
	for i := 0; i < length; i++ {
		b, err := reader.ReadByte()

		if err == io.EOF {
			return Value{}, IncompleteRESPError{
				Reader:  reader,
				Message: fmt.Sprintf("Expected %d bytes of data in bulk string, got %d", length, i),
			}
		}

		bytes.WriteByte(b)
	}

	// Read the \r\n at the end of the bulk string
	if err := readCRLF(reader, "at the end of a bulk string"); err != nil {
		return Value{}, err
	}

	return NewBulkStringValue(bytes.String()), nil
}
