package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type BulkStringAssertion struct {
	ExpectedValue string
}

func NewBulkStringAssertion(expectedValue string) RESPAssertion {
	return BulkStringAssertion{ExpectedValue: expectedValue}
}

func (a BulkStringAssertion) Run(value resp_value.Value) error {
	// Frequently occurs in user submissions
	if value.Type == resp_value.SIMPLE_STRING && value.String() == a.ExpectedValue {
		return fmt.Errorf(
			"Expected bulk string \"%s\", got simple string \"%s\" instead",
			value.String(),
			value.String(),
		)
	}

	bulkStringTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.BULK_STRING}

	if err := bulkStringTypeAssertion.Run(value); err != nil {
		return err
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
