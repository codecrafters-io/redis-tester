package inspectable_byte_string

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormattedString(t *testing.T) {
	bytes := []byte("+OK\r\n")
	ibs := NewInspectableByteString(bytes)
	assert.Equal(t, `"+OK\r\n"`, ibs.FormattedString())
}

func TestGetOffsetInFormattedString(t *testing.T) {
	bytes := []byte("+OK\r\n")
	ibs := NewInspectableByteString(bytes)

	assert.Equal(t, 1, ibs.GetOffsetInFormattedString(0))
	assert.Equal(t, 2, ibs.GetOffsetInFormattedString(1))
	assert.Equal(t, 3, ibs.GetOffsetInFormattedString(2))
	assert.Equal(t, 4, ibs.GetOffsetInFormattedString(3))
	assert.Equal(t, 6, ibs.GetOffsetInFormattedString(4))
}

func TestTruncateAroundOffset(t *testing.T) {
	bytes := []byte("+OK\r\n")
	ibs := NewInspectableByteString(bytes)

	assert.Equal(t, `"+OK\r\n"`, ibs.TruncateAroundOffset(4).FormattedString())
	assert.Equal(t, `"+OK\r\n"`, ibs.TruncateAroundOffset(5).FormattedString())
	assert.Equal(t, `"+OK\r\n"`, ibs.TruncateAroundOffset(6).FormattedString())

	bytes = []byte{}

	for i := 0; i < 10; i++ {
		bytes = append(bytes, []byte(fmt.Sprintf("helloworld%d", i))...)
	}

	ibs = NewInspectableByteString(bytes)
	assert.Equal(t, `"rld3helloworld4helloworld5hell"`, ibs.TruncateAroundOffset(60).FormattedString())
}

func TestFormatWithHighlightedOffset(t *testing.T) {
	bytes := []byte("+OK\r\n")
	ibs := NewInspectableByteString(bytes)
	highlightOffset := 4
	highlightText := "error"
	formattedStringPrefix := "Received: "

	expected := strings.TrimSpace(`
Received: "+OK\r\n" (suffix)
                ^ error
	`)
	result := ibs.FormatWithHighlightedOffset(highlightOffset, highlightText, formattedStringPrefix, " (suffix)")

	assert.Equal(t, expected, result)
}
