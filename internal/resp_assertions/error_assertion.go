package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ErrorAssertion struct {
	ExpectedValue string
}

func NewErrorAssertion(expectedValue string) RESPAssertion {
	return ErrorAssertion{ExpectedValue: expectedValue}
}

func (a ErrorAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ERROR {
		return fmt.Errorf("Expected error, got %s", value.Type)
	}

	if value.Error() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.Error())
	}

	return nil
}
