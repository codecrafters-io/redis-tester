package resp_assertions

import (
	"fmt"
	"sort"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ArrayAssertion struct {
	Elements                   []string
	shouldSortBeforeComparison bool
}

func NewArrayAssertion(elements []string, shouldSortBeforeComparison bool) RESPAssertion {
	return ArrayAssertion{
		Elements:                   elements,
		shouldSortBeforeComparison: shouldSortBeforeComparison,
	}
}

func (a ArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array type, got %s", value.Type)
	}

	response := value.Array()

	if len(response) < 1 {
		return fmt.Errorf("Expected array with at least 1 element, got %d elements", len(response))
	}

	if len(response) != len(a.Elements) {
		return fmt.Errorf("Expected command to have %d arguments, got %d", len(a.Elements), len(response))
	}

	sortedStringsFromResponse := []string{}

	for i, value := range response {
		if value.Type != resp_value.SIMPLE_STRING && value.Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected argument %d to be a string, got %s", i+1, value.Type)
		}
		sortedStringsFromResponse = append(sortedStringsFromResponse, value.String())
	}

	sort.Strings(sortedStringsFromResponse)
	sort.Strings(a.Elements)

	for i, expectedArg := range a.Elements {
		value := sortedStringsFromResponse[i]

		if value != expectedArg {
			return fmt.Errorf("Expected argument #%d to be %q, got %q", i+1, expectedArg, value)
		}
	}

	return nil
}
