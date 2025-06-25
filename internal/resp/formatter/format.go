package formatter

import "github.com/tidwall/pretty"

var defaultOptions = &pretty.Options{
	Indent: "  ",
	Width:  32,
}

func Prettify(json []byte) string {
	return string(pretty.PrettyOptions(json, defaultOptions))
}
