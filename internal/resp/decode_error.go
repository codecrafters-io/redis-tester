package resp

import (
	"bytes"
	"fmt"
	"strings"

	inspectable_byte_string "github.com/codecrafters-io/redis-tester/internal/inspectable_byte_string"
)

type IncompleteRESPError struct {
	Reader  *bytes.Reader
	Message string
}

type InvalidRESPError struct {
	Reader  *bytes.Reader
	Message string
}

func (e IncompleteRESPError) Error() string {
	return formatDetailedError(e.Reader, e.Message)
}

func (e InvalidRESPError) Error() string {
	return formatDetailedError(e.Reader, e.Message)
}

func getReaderOffset(reader *bytes.Reader) int {
	return int(reader.Size()) - reader.Len()
}

func readBytesFromReader(reader *bytes.Reader) []byte {
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

func formatDetailedError(reader *bytes.Reader, message string) string {
	lines := []string{}

	offset := getReaderOffset(reader)
	receivedBytes := readBytesFromReader(reader)
	receivedByteString := inspectable_byte_string.NewInspectableByteString(receivedBytes)
	receivedByteString = receivedByteString.TruncateAroundOffset(offset)

	lines = append(lines, fmt.Sprintf("Received: %s", receivedByteString.FormattedString()))
	lines = append(lines, offsetPointerString(len("Received: ")+receivedByteString.GetOffsetInFormattedString(offset)))
	lines = append(lines, fmt.Sprintf("Error: %s", message))

	return strings.Join(lines, "\n")
}

func offsetPointerString(offset int) string {
	return strings.Repeat(" ", offset) + "^ error"
}
