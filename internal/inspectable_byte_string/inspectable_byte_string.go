package inspectable_byte_string

import "fmt"

type InspectableByteString struct {
	bytes []byte
}

func NewInspectableByteString(bytes []byte) *InspectableByteString {
	return &InspectableByteString{bytes: bytes}
}

// GetOffsetInFormattedString returns a string that represents the byteOffset in the formatted string
//
// For example:
//   - If the string is "+OK\r\n"
//   - And byteOffset is 4 (i.e. \n, the 5th byte)
//   - The return value will be 6 (i.e. the 6th character in the formatted string)
func (s *InspectableByteString) GetOffsetInFormattedString(byteOffset int) int {
	formattedBytesBefore := fmt.Sprintf("%q", string(s.bytes[:byteOffset]))
	return len(formattedBytesBefore) - 1
}

func (s *InspectableByteString) FormattedString() string {
	return fmt.Sprintf("%q", string(s.bytes))
}
