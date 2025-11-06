package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// OrderedBulkStringArrayAssertion : Order of the actual and expected values matters.
// We don't alter the ordering.
type OrderedBulkStringArrayAssertion struct {
	ExpectedValue []string
}

func NewOrderedBulkStringArrayAssertion(expectedValue []string) RESPAssertion {
	return OrderedBulkStringArrayAssertion{ExpectedValue: expectedValue}
}

func (a OrderedBulkStringArrayAssertion) Run(value resp_value.Value) error {
	arrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := arrayTypeAssertion.Run(value); err != nil {
		return err
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	for i, expectedValue := range a.ExpectedValue {
		actualElement := value.Array()[i]

		if actualElement.Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected element #%d to be a bulk string, got %s", i+1, actualElement.Type)
		}

		if actualElement.String() != expectedValue {
			return fmt.Errorf("Expected element #%d to be %q, got %q", i+1, expectedValue, actualElement.String())
		}
	}

	return nil
}
