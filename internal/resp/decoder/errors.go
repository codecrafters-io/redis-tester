package resp_decoder

import (
	"bytes"
	"fmt"
	"strings"

	inspectable_byte_string "github.com/codecrafters-io/tester-utils/inspectable_byte_string"
)

type IncompleteInputError struct {
	Reader  *bytes.Reader
	Message string
}

type InvalidInputError struct {
	Reader  *bytes.Reader
	Message string
}

func (e IncompleteInputError) Error() string {
	return formatDetailedError(e.Reader, e.Message)
}

func (e InvalidInputError) Error() string {
	return formatDetailedError(e.Reader, e.Message)
}

func formatDetailedError(reader *bytes.Reader, message string) string {
	lines := []string{}

	offset := getReaderOffset(reader)
	receivedBytes := readBytesFromReader(reader)
	receivedByteString := inspectable_byte_string.NewInspectableByteString(receivedBytes)

	suffix := ""

	if len(receivedBytes) == 0 {
		suffix = " (no content received)"
	}

	lines = append(lines, receivedByteString.FormatWithHighlightedOffset(offset, "error", "Received: ", suffix))
	lines = append(lines, fmt.Sprintf("Error: %s", message))

	return strings.Join(lines, "\n")
}

func getReaderOffset(reader *bytes.Reader) int {
	return int(reader.Size()) - reader.Len()
}

func readBytesFromReader(reader *bytes.Reader) []byte {
	previousOffset := getReaderOffset(reader)
	defer reader.Seek(int64(previousOffset), 0)

	reader.Seek(0, 0)
	bytes := make([]byte, reader.Len())

	if reader.Len() == 0 {
		return bytes
	}

	n, err := reader.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("Error reading from reader: %s", err)) // This should never happen
	}
	if n != len(bytes) {
		panic(fmt.Sprintf("Expected to read %d bytes, but only read %d", len(bytes), n)) // This should never happen
	}

	return bytes
}
