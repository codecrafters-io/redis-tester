package resp_assertions

import (
	"fmt"
	"sort"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StringArrayAssertion struct {
	ExpectedValue              []string
	shouldSortBeforeComparison bool
}

func NewStringArrayAssertion(expectedValue []string, shouldSortbeforeComparison bool) RESPAssertion {
	return StringArrayAssertion{ExpectedValue: expectedValue, shouldSortBeforeComparison: shouldSortbeforeComparison}
}

func (a StringArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	actualElementStringArray := make([]string, len(value.Array()))
	for i := range value.Array() {
		actualElement := value.Array()[i]
		if actualElement.Type != resp_value.BULK_STRING && actualElement.Type != resp_value.SIMPLE_STRING {
			return fmt.Errorf("Expected element #%d to be a string, got %s", i+1, actualElement.Type)
		}
		actualElementStringArray[i] = value.Array()[i].String()
	}

	if a.shouldSortBeforeComparison {
		sort.Strings(actualElementStringArray)
		sort.Strings(a.ExpectedValue)
	}

	for i, expectedValue := range a.ExpectedValue {
		actualElement := actualElementStringArray[i]

		if actualElement != expectedValue {
			return fmt.Errorf("Expected element #%d to be %q, got %q", i+1, expectedValue, actualElement)
		}
	}

	return nil
}
