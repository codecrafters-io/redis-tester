package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

func DecodeFullResyncRDBFile(data []byte) (rdbFileContents []byte, readBytesCount int, err error) {
	reader := bytes.NewReader(data)

	rdbFileContents, err = doDecodeFullResyncRDBFile(reader)
	if err != nil {
		return nil, 0, err
	}

	return rdbFileContents, len(data) - reader.Len(), nil
}

func doDecodeFullResyncRDBFile(reader *bytes.Reader) ([]byte, error) {
	firstByte, err := reader.ReadByte()
	if err == io.EOF {
		return nil, IncompleteInputError{
			Reader:  reader,
			Message: "Expected first byte of RDB file message to be $",
		}
	}

	if firstByte != '$' {
		reader.UnreadByte() // Ensure the error points to the correct byte

		return nil, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Expected first byte of RDB file message to be $, got %q", string(firstByte)),
		}
	}

	offsetBeforeLength := getReaderOffset(reader)
	lengthBytes, err := readUntilCRLF(reader)
	if err == io.EOF {
		return nil, IncompleteInputError{
			Reader:  reader,
			Message: `Expected \r\n after RDB file length`,
		}
	}

	length, err := strconv.Atoi(string(lengthBytes))
	if err != nil {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return nil, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid RDB file length: %q, expected a number", string(lengthBytes)),
		}
	}

	if length < 1 {
		// Ensure error points to the correct byte
		reader.Seek(int64(offsetBeforeLength), io.SeekStart)

		return nil, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("Invalid RDB file length: %d, expected a positive integer", length),
		}
	}

	bytes := bytes.NewBuffer([]byte{})
	for i := 0; i < length; i++ {
		b, err := reader.ReadByte()

		if err == io.EOF {
			return nil, IncompleteInputError{
				Reader:  reader,
				Message: fmt.Sprintf("Expected %d bytes of data in RDB file message, got %d", length, i),
			}
		}

		bytes.WriteByte(b)
	}

	return bytes.Bytes(), nil

}
