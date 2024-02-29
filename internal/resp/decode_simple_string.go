package resp

import (
	"bytes"
	"io"
)

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
