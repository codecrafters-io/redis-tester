package resp_assertions

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/resp"
)

type RESPAssertion interface {
	Run(value resp.Value) error
}

type StringAssertion struct {
	ExpectedValue string
}

func NewStringValueAssertion(expectedValue string) RESPAssertion {
	return StringAssertion{ExpectedValue: expectedValue}
}

func (a StringAssertion) Run(value resp.Value) error {
	if value.Type != resp.SIMPLE_STRING && value.Type != resp.BULK_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", value.Type)
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
