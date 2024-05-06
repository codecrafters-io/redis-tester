package resp_assertions

import (
	"encoding/json"
	"fmt"
	"sort"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// UnorderedStringArrayAssertion : Order of the actual and expected values doesn't matter.
// We sort the expected and actual values before comparing them.
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

	actualElementStringArray := make([]string, len(value.Array()))
	for i := range value.Array() {
		actualElement := value.Array()[i]
		if actualElement.Type != resp_value.BULK_STRING && actualElement.Type != resp_value.SIMPLE_STRING {
			return fmt.Errorf("Expected element #%d to be a string, got %s", i+1, actualElement.Type)
		}
		actualElementStringArray[i] = value.Array()[i].String()
	}

	expectedValueArrayForPrinting, _ := json.Marshal(a.ExpectedValue)
	expectedValueStringArray := make([]string, len(a.ExpectedValue))
	copy(expectedValueStringArray, a.ExpectedValue)
	sort.Strings(actualElementStringArray)
	sort.Strings(expectedValueStringArray)

	for i, expectedValue := range expectedValueStringArray {
		actualElement := actualElementStringArray[i]

		if actualElement != expectedValue {
			return fmt.Errorf("Expected: %v (in any order), got %v", string(expectedValueArrayForPrinting), value.FormattedString())
		}
	}

	return nil
}
