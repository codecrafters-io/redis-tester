package resp_assertions

import (
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RegexErrorAssertion struct {
	ExpectedPattern *regexp.Regexp
}

func NewRegexErrorAssertion(expectedPattern string) RESPAssertion {
	pattern, err := regexp.Compile(expectedPattern)
	if err != nil {
		panic(err)
	}
	return RegexErrorAssertion{ExpectedPattern: pattern}
}

func (a RegexErrorAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ERROR}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	if !a.ExpectedPattern.MatchString(value.Error()) {
		return fmt.Errorf("Expected error to match (%q), got (%q)", a.ExpectedPattern.String(), value.Error())
	}

	return nil
}
