package inspectable_byte_string

import (
	"fmt"
	"strings"
)

type InspectableByteString struct {
	bytes []byte

	truncationStartIndex int
}

func NewInspectableByteString(bytes []byte) InspectableByteString {
	return InspectableByteString{bytes: bytes}
}

// FormatWithHighlightedOffset returns a string that represents the bytes with the byteOffset highlighted
//
// For example, if called with highlightOffset 4, highlightText "error" and formattedString "Received: ", the return value will be:
//
// > Received: "+OK\r\n"
// >                 ^ error
func (s InspectableByteString) FormatWithHighlightedOffset(highlightOffset int, highlightText string, formattedStringPrefix string, formattedStringSuffix string) string {
	s = s.TruncateAroundOffset(highlightOffset)

	lines := []string{}

	lines = append(lines, fmt.Sprintf("%s%s%s", formattedStringPrefix, s.FormattedString(), formattedStringSuffix))

	offsetPointerLine := ""
	offsetPointerLine += strings.Repeat(" ", len(formattedStringPrefix)+s.GetOffsetInFormattedString(highlightOffset))
	offsetPointerLine += "^ " + highlightText
	lines = append(lines, offsetPointerLine)

	return strings.Join(lines, "\n")
}

func (s InspectableByteString) FormattedString() string {
	return fmt.Sprintf("%q", string(s.bytes))
}

func (s InspectableByteString) TruncateAroundOffset(offset int) InspectableByteString {
	// We've got about 50 characters to use in the terminal line.
	start := max(0, offset-20)
	end := max(0, min(len(s.bytes), start+30))

	return InspectableByteString{
		bytes:                s.bytes[start:end],
		truncationStartIndex: start,
	}
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
