package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StringPresentInArrayAssertion struct {
	expectedString string
}

func NewStringPresentInArrayAssertion(expectedElement string) *StringPresentInArrayAssertion {
	return &StringPresentInArrayAssertion{
		expectedString: expectedElement,
	}
}

func (a StringPresentInArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	found := false

	for _, actualElement := range value.Array() {

		if actualElement.String() == a.expectedString {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("'%s' not found in the array", a.expectedString)
	}

	return nil
}
