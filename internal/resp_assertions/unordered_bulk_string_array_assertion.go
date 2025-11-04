package resp_assertions

import (
	"encoding/json"
	"fmt"
	"sort"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// UnorderedBulkStringArrayAssertion : Order of the actual and expected values doesn't matter.
// We sort the expected and actual values before comparing them.
type UnorderedBulkStringArrayAssertion struct {
	ExpectedValue []string
}

func NewUnorderedBulkStringArrayAssertion(expectedValue []string) RESPAssertion {
	return UnorderedBulkStringArrayAssertion{ExpectedValue: expectedValue}
}

func (a UnorderedBulkStringArrayAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	if len(value.Array()) != len(a.ExpectedValue) {
		return fmt.Errorf("Expected %d elements in array, got %d (%s)", len(a.ExpectedValue), len(value.Array()), value.FormattedString())
	}

	actualElementStringArray := make([]string, len(value.Array()))
	for i := range value.Array() {
		actualElement := value.Array()[i]

		if actualElement.Type != resp_value.BULK_STRING {
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
