package instrumented_resp_connection

import (
	"fmt"
	"strings"
)

// quoteIfNeeded quotes a string if it contains escapable characters or spaces
func quoteIfNeeded(s string) string {
	quoted := fmt.Sprintf("%q", s)
	trimmedQuotes := strings.Trim(quoted, "\"")

	// if the string does not change, no escapable chracter was present
	if s == trimmedQuotes && !strings.Contains(s, " ") {
		return s
	}

	return quoted
}

// quoteCLIArgs applies quoteIfNeeded to each CLI argument
func quoteCLIArgs(args []string) []string {
	result := make([]string, len(args))
	for i, a := range args {
		result[i] = quoteIfNeeded(a)
	}
	return result
}
