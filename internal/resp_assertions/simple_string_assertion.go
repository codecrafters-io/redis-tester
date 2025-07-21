package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type SimpleStringAssertion struct {
	ExpectedValue string
}

func NewSimpleStringAssertion(expectedValue string) RESPAssertion {
	return SimpleStringAssertion{ExpectedValue: expectedValue}
}

func (a SimpleStringAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.SIMPLE_STRING {
		return fmt.Errorf("Expected simple string, got %s", value.Type)
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
