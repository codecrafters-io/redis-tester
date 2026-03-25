package resp_decoder

import (
	"bytes"
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// DecodeCommandsFromAppendOnlyFile parses Redis AOF-style content: one RESP array
// (typically bulk strings per argument) per command, concatenated with no separator.
func DecodeCommandsFromAppendOnlyFile(data []byte) ([][]string, error) {
	reader := bytes.NewReader(data)
	var commands [][]string

	for reader.Len() > 0 {
		v, err := doDecodeValue(reader)

		if err != nil {
			return commands, err
		}

		if v.Type != resp_value.ARRAY {
			return commands, InvalidInputError{
				Reader:  reader,
				Message: fmt.Sprintf("Expected RESP array for AOF command, got %q", v.Type),
			}
		}

		args, err := aofArrayToStringsArray(reader, v)

		if err != nil {
			return commands, err
		}

		commands = append(commands, args)
	}

	return commands, nil
}

func aofArrayToStringsArray(reader *bytes.Reader, v resp_value.Value) ([]string, error) {
	elems := v.Array()
	out := make([]string, 0, len(elems))

	for _, elem := range elems {
		if elem.Type != resp_value.BULK_STRING {
			return nil, InvalidInputError{
				Reader: reader,
				Message: fmt.Sprintf(
					"AOF command must be a RESP array of bulk strings, got element type %q",
					elem.Type,
				),
			}
		}

		out = append(out, elem.String())
	}

	return out, nil
}
