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
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.BULK_STRING}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
