package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type BulkStringAbsentFromArrayAssertion struct {
	StringExpectedToBeAbsent string
}

func (a BulkStringAbsentFromArrayAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	array := value.Array()

	for _, element := range array {
		if element.Type == resp_value.BULK_STRING && element.String() == a.StringExpectedToBeAbsent {
			return fmt.Errorf("Expected string '%s' to be absent from the array, but is present", a.StringExpectedToBeAbsent)
		}
	}

	return nil
}
