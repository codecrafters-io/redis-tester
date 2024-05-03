package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type UnorderedStringArrayAssertion struct {
	ExpectedValue []string
}

func NewUnorderedStringArrayAssertion(expectedValue []string) RESPAssertion {
	return UnorderedStringArrayAssertion{ExpectedValue: expectedValue}
}

func (a UnorderedStringArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	for i, expectedValue := range a.ExpectedValue {
		actualElement := value.Array()[i]

		if actualElement.Type != resp_value.BULK_STRING && actualElement.Type != resp_value.SIMPLE_STRING {
			return fmt.Errorf("Expected element #%d to be a string, got %s", i+1, actualElement.Type)
		}

		if actualElement.String() != expectedValue {
			return fmt.Errorf("Expected element #%d to be %q, got %q", i+1, expectedValue, actualElement.String())
		}
	}

	return nil
}
