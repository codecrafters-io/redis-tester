package resp_assertions

import (
	"bytes"
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// OrderedArrayAssertion : Order of the actual and expected values matters.
// All RESP values are accepted as elements in this array.
// We don't alter the ordering.
type OrderedArrayAssertion struct {
	ExpectedValue []resp_value.Value
}

func NewOrderedArrayAssertion(expectedValue []resp_value.Value) RESPAssertion {
	return OrderedArrayAssertion{ExpectedValue: expectedValue}
}

func (a OrderedArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	for i, expectedValue := range a.ExpectedValue {
		actualElement := value.Array()[i]

		if actualElement.Type != expectedValue.Type {
			return fmt.Errorf("Expected element #%d to be a %s, got %s", i+1, expectedValue.Type, actualElement.Type)
		}

		if expectedValue.Bytes() == nil {
			// This should never happen, but just in case
			// This is the case for ArrayValues
			return fmt.Errorf("CodeCrafters internal error. expectedValue Bytes of type: %s is nil", expectedValue.Type)
		}

		// ToDo: Equal or EqualFold ?
		if !bytes.Equal(actualElement.Bytes(), expectedValue.Bytes()) {
			return fmt.Errorf("Expected element #%d to be %s, got %s", i+1, expectedValue.FormattedString(), actualElement.FormattedString())
		}
	}

	return nil
}
