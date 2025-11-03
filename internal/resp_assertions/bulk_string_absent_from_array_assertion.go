package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type BulkStringAbsentFromArrayAssertion struct {
	StringExpectedToBeAbsent string
}

func (a BulkStringAbsentFromArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	array := value.Array()

	for _, element := range array {
		if element.Type == resp_value.BULK_STRING && element.String() == a.StringExpectedToBeAbsent {
			return fmt.Errorf("Expected string '%s' to be absent from the array, but is present", a.StringExpectedToBeAbsent)
		}
	}

	return nil
}
