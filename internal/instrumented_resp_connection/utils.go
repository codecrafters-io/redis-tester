package instrumented_resp_connection

import (
	"fmt"
	"strings"
)

// quoteIfHasSpaceOrEscapeSequence quotes a string if it contains escapable characters or spaces
func quoteIfHasSpaceOrEscapeSequence(s string) string {
	quoted := fmt.Sprintf("%q", s)
	trimmedQuotes := strings.Trim(quoted, "\"")

	// if the string does not change, no escapable chracter was present
	if s == trimmedQuotes && !strings.Contains(s, " ") {
		return s
	}

	return quoted
}

// quoteCLICommand applies quoteIfHasSpaceOrEscapeSequence to each CLI argument
func quoteCLICommand(commandWithArgs []string) []string {
	result := make([]string, len(commandWithArgs))
	for i, a := range commandWithArgs {
		result[i] = quoteIfHasSpaceOrEscapeSequence(a)
	}
	return result
}
