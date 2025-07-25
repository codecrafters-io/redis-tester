package instrumented_resp_connection

import (
	"fmt"
	"strings"
)

// quoteCLIArg quotes a string if it contains an escape-able character or a space
func quoteCLIArg(s string) string {
	quoted := fmt.Sprintf("%q", s)
	trimmedQuotes := strings.Trim(quoted, "\"")

	// if the string does not change, no escape-able chracter was present
	if s == trimmedQuotes && !strings.Contains(s, " ") {
		return s
	}

	return quoted
}
