package resp_decoder

import (
	"bytes"
	"fmt"
	"io"
)

func readUntilCRLF(r *bytes.Reader) ([]byte, error) {
	return readUntil(r, []byte("\r\n"))
}

func readCRLF(reader *bytes.Reader, locationDescriptor string) (err error) {
	errorMessage := fmt.Sprintf(`Expected \r\n %s`, locationDescriptor)

	b, err := reader.ReadByte()
	if err == io.EOF {
		return IncompleteInputError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	if b != '\r' {
		return InvalidInputError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	b, err = reader.ReadByte()
	if err == io.EOF {
		return IncompleteInputError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	if b != '\n' {
		return InvalidInputError{
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
