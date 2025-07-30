package instrumented_resp_connection

import (
	"fmt"
	"strings"
)

// quoteIfHasSpaceOrEscapeSequence quotes a string if it contains escapable characters or spaces
func quoteIfHasSpaceOrEscapeSequence(s string) string {
	quoted := fmt.Sprintf("%q", s)

	// if the string does not change on trimming double quotes, no escapable chracter was present
	if (s == strings.Trim(quoted, "\"")) && (!strings.Contains(s, " ")) {
		return s
	}

	return quoted
}

// quoteCLICommand applies quoteIfHasSpaceOrEscapeSequence to each argument in commandWithArgs,
// then joins them into a single space-separated string.
func quoteCLICommand(commandWithArgs []string) string {
	result := make([]string, len(commandWithArgs))
	for i, a := range commandWithArgs {
		result[i] = quoteIfHasSpaceOrEscapeSequence(a)
	}
	return strings.Join(result, " ")
}
