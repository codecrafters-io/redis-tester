package resp_assertions

import (
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ErrorRegexAssertion struct {
	expectedRegex regexp.Regexp
}

func NewErrorRegexAssertion(expectedRegex regexp.Regexp) RESPAssertion {
	return ErrorRegexAssertion{expectedRegex: expectedRegex}
}

func (a ErrorRegexAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ERROR {
		return fmt.Errorf("Expected error, got %s", value.Type)
	}

	if !a.expectedRegex.MatchString(value.Error()) {
		return fmt.Errorf("Expected error to match (%q), got (%q)", a.expectedRegex.String(), value.Error())
	}

	return nil
}
