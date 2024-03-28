package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type IntegerAssertion struct {
	ExpectedValue int
}

func NewIntegerAssertion(expectedValue int) RESPAssertion {
	return IntegerAssertion{ExpectedValue: expectedValue}
}

func (a IntegerAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.INTEGER {
		return fmt.Errorf("Expected integer, got %s", value.Type)
	}

	if value.Integer() != a.ExpectedValue {
		return fmt.Errorf("Expected %d, got %d", a.ExpectedValue, value.Integer())
	}

	return nil
}
