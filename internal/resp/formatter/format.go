package formatter

import (
	"bytes"

	"github.com/tidwall/pretty"
)

var defaultOptions = &pretty.Options{
	Indent: "  ",
	Width:  32,
}

func Prettify(json []byte) string {
	prettified := pretty.PrettyOptions(json, defaultOptions)
	prettified = bytes.TrimRight(prettified, "\n")
	return string(prettified)
}
