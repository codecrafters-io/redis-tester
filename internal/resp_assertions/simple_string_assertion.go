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
	// Frequently occurs in user submissions
	if value.Type == resp_value.BULK_STRING && value.String() == a.ExpectedValue {
		return fmt.Errorf(
			"Expected simple string \"%s\", got bulk string \"%s\" instead",
			value.String(),
			value.String(),
		)
	}

	simpleStringTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.SIMPLE_STRING}

	if err := simpleStringTypeAssertion.Run(value); err != nil {
		return err
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
