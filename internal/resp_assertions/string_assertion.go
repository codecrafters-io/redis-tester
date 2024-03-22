package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StringAssertion struct {
	ExpectedValue string
}

func NewStringAssertion(expectedValue string) RESPAssertion {
	return StringAssertion{ExpectedValue: expectedValue}
}

func (a StringAssertion) Run(value resp_value.Value) RESPAssertionResult {
	if value.Type != resp_value.SIMPLE_STRING && value.Type != resp_value.BULK_STRING {
		return RESPAssertionResult{
			ErrorMessages: []string{fmt.Sprintf("Expected simple string or bulk string, got %s", value.Type)},
		}
	}

	if value.String() != a.ExpectedValue {
		return RESPAssertionResult{
			ErrorMessages: []string{fmt.Sprintf("Expected %q, got %q", a.ExpectedValue, value.String())},
		}
	}

	return RESPAssertionResult{SuccessMessages: []string{fmt.Sprintf("Received %q", value.String())}}
}
