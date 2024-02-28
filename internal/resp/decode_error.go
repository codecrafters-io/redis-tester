package resp

import (
	"bytes"
	"fmt"
	"strings"
)

type IncompleteRESPError struct {
	Reader  *bytes.Reader
	Message string
}

type InvalidRESPError struct {
	Reader  *bytes.Reader
	Message string
}

func (e IncompleteRESPError) DetailedError() string {
	lines := []string{}

	receivedBytes := readBytesFromReader(e.Reader)

	lines = append(lines, fmt.Sprintf("Received: %q", string(receivedBytes)))
	lines = append(lines, e.Message)

	return strings.Join(lines, "\n")
}

func (e IncompleteRESPError) Error() string {
	// TODO: Make this more readable
	return fmt.Sprintf("Incomplete RESP: %s.", e.Message)
}

func (e InvalidRESPError) DetailedError() string {
	lines := []string{}

	receivedBytes := readBytesFromReader(e.Reader)

	lines = append(lines, fmt.Sprintf("Received: %q", string(receivedBytes)))
	lines = append(lines, e.Message)

	return strings.Join(lines, "\n")
}

func (e InvalidRESPError) Error() string {
	return fmt.Sprintf("Invalid RESP: %s", e.Message)
}

func getReaderOffset(reader *bytes.Reader) int {
	return int(reader.Size()) - reader.Len()
}

func readBytesFromReader(reader *bytes.Reader) []byte {
	reader.Seek(0, 0)
	bytes := make([]byte, reader.Len())
	n, err := reader.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("Error reading from reader: %s", err)) // This should never happen
	}
	if n != len(bytes) {
		panic(fmt.Sprintf("Expected to read %d bytes, but only read %d", len(bytes), n)) // This should never happen
	}

	return bytes
}
