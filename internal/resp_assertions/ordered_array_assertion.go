package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// OrderedArrayAssertion : Order of the actual and expected values matters.
// All RESP values are accepted as elements in this array.
// We don't alter the ordering.
// For each element in the array, we run the corresponding assertion.
type OrderedArrayAssertion struct {
	ExpectedValue []RESPAssertion
}

func NewOrderedArrayAssertion(expectedValue []RESPAssertion) RESPAssertion {
	return OrderedArrayAssertion{ExpectedValue: expectedValue}
}

func (a OrderedArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	for i, assertion := range a.ExpectedValue {
		actualElement := value.Array()[i]

		if err := assertion.Run(actualElement); err != nil {
			return err
		}
	}

	return nil
}
