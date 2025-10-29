package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StringAbsentInArrayAssertion struct {
	expectedAbsentElement string
}

func NewStringAbsentInArrayAssertion(expectedElement string) *StringAbsentInArrayAssertion {
	return &StringAbsentInArrayAssertion{
		expectedAbsentElement: expectedElement,
	}
}

func (a StringAbsentInArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected an array, got %s", value.Type)
	}

	for _, actualElement := range value.Array() {

		if actualElement.String() == a.expectedAbsentElement {
			return fmt.Errorf("Expecting '%s' to be absent from the array, but is present", a.expectedAbsentElement)
		}
	}

	return nil
}
