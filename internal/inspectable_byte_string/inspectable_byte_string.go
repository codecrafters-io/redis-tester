package inspectable_byte_string

import (
	"fmt"
)

type InspectableByteString struct {
	bytes []byte

	truncationStartIndex int
}

func NewInspectableByteString(bytes []byte) InspectableByteString {
	return InspectableByteString{bytes: bytes}
}

// GetOffsetInFormattedString returns a string that represents the byteOffset in the formatted string
//
// For example:
//   - If the string is "+OK\r\n"
//   - And byteOffset is 4 (i.e. \n, the 5th byte)
//   - The return value will be 6 (i.e. the 6th character in the formatted string)
func (s InspectableByteString) GetOffsetInFormattedString(byteOffset int) int {
	if s.truncationStartIndex != 0 {
		byteOffset = byteOffset - s.truncationStartIndex
	}

	formattedBytesBefore := fmt.Sprintf("%q", string(s.bytes[:byteOffset]))
	return len(formattedBytesBefore) - 1
}

func (s InspectableByteString) FormattedString() string {
	return fmt.Sprintf("%q", string(s.bytes))
}

func (s InspectableByteString) TruncateAroundOffset(offset int) InspectableByteString {
	// We've got about 50 characters to use in the terminal line.
	start := max(0, offset-20)
	end := max(0, min(len(s.bytes), offset+10))

	return InspectableByteString{
		bytes:                s.bytes[start:end],
		truncationStartIndex: start,
	}
}
