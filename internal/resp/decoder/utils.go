package resp_decoder

import (
	"bytes"
	"io"
)

func readUntilCRLF(r *bytes.Reader) ([]byte, error) {
	return readUntil(r, []byte("\r\n"))
}

func readCRLF(reader *bytes.Reader, errorMessage string) (err error) {
	offsetBeforeCRLF := getReaderOffset(reader)

	b, err := reader.ReadByte()
	if err == io.EOF {
		return IncompleteInputError{
			Reader:  reader,
			Message: errorMessage,
		}
	}

	if b != '\r' {
		reader.Seek(int64(offsetBeforeCRLF), io.SeekStart)

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
		reader.Seek(int64(offsetBeforeCRLF), io.SeekStart)

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
