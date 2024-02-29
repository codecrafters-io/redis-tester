package resp_assertions

import (
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RegexStringAssertion struct {
	ExpectedPattern *regexp.Regexp
}

func NewRegexStringAssertion(expectedPattern string) RESPAssertion {
	pattern, err := regexp.Compile(expectedPattern)
	if err != nil {
		panic(err)
	}

	return RegexStringAssertion{ExpectedPattern: pattern}
}

func (a RegexStringAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.SIMPLE_STRING && value.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", value.Type)
	}

	if !a.ExpectedPattern.MatchString(value.String()) {
		return fmt.Errorf("Expected %q to match the pattern %q.", value.String(), a.ExpectedPattern.String())
	}

	return nil
}
