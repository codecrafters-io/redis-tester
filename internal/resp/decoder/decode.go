package resp_decoder

import (
	"bytes"
	"fmt"
	"io"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func Decode(data []byte) (value resp_value.Value, readBytesCount int, err error) {
	reader := bytes.NewReader(data)

	value, err = doDecodeValue(reader)
	if err != nil {
		return resp_value.Value{}, 0, err
	}

	return value, len(data) - reader.Len(), nil
}

func doDecodeValue(reader *bytes.Reader) (resp_value.Value, error) {
	firstByte, err := reader.ReadByte()
	if err == io.EOF {
		return resp_value.Value{}, IncompleteRESPError{
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

		return resp_value.Value{}, InvalidRESPError{
			Reader:  reader,
			Message: fmt.Sprintf("%q is not a valid start of a RESP value (expected +, -, :, $ or *)", string(firstByte)),
		}
	}
}
