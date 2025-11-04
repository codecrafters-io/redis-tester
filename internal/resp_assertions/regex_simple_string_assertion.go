package resp_assertions

import (
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RegexSimpleStringAssertion struct {
	ExpectedPattern *regexp.Regexp
}

func NewRegexSimpleStringAssertion(expectedPattern string) RESPAssertion {
	pattern, err := regexp.Compile(expectedPattern)
	if err != nil {
		panic(err)
	}

	return RegexSimpleStringAssertion{ExpectedPattern: pattern}
}

func (a RegexSimpleStringAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.SIMPLE_STRING}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	if !a.ExpectedPattern.MatchString(value.String()) {
		return fmt.Errorf("Expected %q to match the pattern %q.", value.String(), a.ExpectedPattern.String())
	}

	return nil
}
