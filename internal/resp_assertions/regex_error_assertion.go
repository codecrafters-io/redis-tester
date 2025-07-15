package resp_assertions

import (
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RegexErrorAssertion struct {
	expectedRegex *regexp.Regexp
}

func NewErrorRegexAssertion(expectedPattern string) RESPAssertion {
	pattern, err := regexp.Compile(expectedPattern)
	if err != nil {
		panic(err)
	}
	return RegexErrorAssertion{expectedRegex: pattern}
}

func (a RegexErrorAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ERROR {
		return fmt.Errorf("Expected error, got %s", value.Type)
	}

	if !a.expectedRegex.MatchString(value.Error()) {
		return fmt.Errorf("Expected error to match (%q), got (%q)", a.expectedRegex.String(), value.Error())
	}

	return nil
}
