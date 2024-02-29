package resp

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

func Decode(data []byte) (value Value, readBytesCount int, err error) {
	reader := bytes.NewReader(data)

	value, err = doDecodeValue(reader)
	if err != nil {
		return Value{}, 0, err
	}

	return value, len(data) - reader.Len(), nil
}

func doDecodeValue(reader *bytes.Reader) (Value, error) {
	firstByte, err := reader.ReadByte()
	if err == io.EOF {
		return Value{}, IncompleteRESPError{
			Reader:  reader,
			Message: "Expected start of a new RESP value (either +, -, :, $ or *)",
		}
	}

	switch firstByte {
	case '+':
		return decodeSimpleString(reader)
	case '$':
		return decodeBulkString(reader)
	// TODO: Implement the rest of the types
	default:
		reader.UnreadByte() // Ensure the error points to the correct byte

		return Value{}, InvalidRESPError{
			Reader:  reader,
			Message: fmt.Sprintf("%q is not a valid start of a RESP value (expected +, -, :, $ or *)", string(firstByte)),
		}
	}
}

func decodeSimpleString(reader *bytes.Reader) (Value, error) {
	bytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return Value{}, IncompleteRESPError{
			Reader:  reader,
			Message: `Expected \r\n at the end of a simple string`,
		}
	}

	return NewSimpleStringValue(string(bytes)), nil
}

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

func readUntilCRLF(r *bytes.Reader) ([]byte, error) {
	return readUntil(r, []byte("\r\n"))
}

func readCRLF(reader *bytes.Reader, locationDescriptor string) (err error) {
	errorMessage := fmt.Sprintf(`Expected \r\n %s`, locationDescriptor)

	b, err := reader.ReadByte()
	if err == io.EOF {
		return IncompleteRESPError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	if b != '\r' {
		return InvalidRESPError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	b, err = reader.ReadByte()
	if err == io.EOF {
		return IncompleteRESPError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	if b != '\n' {
		return InvalidRESPError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	return nil
}

func readUntil(r *bytes.Reader, delim []byte) ([]byte, error) {
	var result []byte

	for {
		b, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				panic("expected error to always be io.EOF")
			}

			return result, io.EOF
		}

		result = append(result, b)

		if len(result) >= len(delim) && bytes.Equal(result[len(result)-len(delim):], delim) {
			return result[:len(result)-len(delim)], nil
		}
	}
}
