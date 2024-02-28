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

	oldOffset := getReaderOffset(e.Reader)
	e.Reader.Seek(0, 0)
	bytes, err := e.Reader.Read(make([]byte, e.Reader.Len()))
	if err != nil {
		panic(fmt.Sprintf("Error reading from reader: %s", err)) // This should never happen
	}
	e.Reader.Seek(int64(oldOffset), 0)

	lines = append(lines, e.Message)
	lines = append(lines, fmt.Sprintf("%q", bytes))

	return strings.Join(lines, "\n")
}

func (e IncompleteRESPError) Error() string {
	// TODO: Make this more readable
	return fmt.Sprintf("Incomplete RESP: %s.", e.Message)
}

func (e InvalidRESPError) DetailedError() string {
	lines := []string{}

	e.Reader.Seek(0, 0)
	bytes := make([]byte, e.Reader.Len())

	n, err := e.Reader.Read(bytes)
	if err != nil || n != len(bytes) {
		panic(fmt.Sprintf("Error reading from reader: %s", err)) // This should never happen
	}

	lines = append(lines, fmt.Sprintf("Received: %q", string(bytes)))
	lines = append(lines, e.Message)

	return strings.Join(lines, "\n")
}

func (e InvalidRESPError) Error() string {
	return fmt.Sprintf("Invalid RESP: %s", e.Message)
}

func getReaderOffset(reader *bytes.Reader) int {
	return int(reader.Size()) - reader.Len()
}
