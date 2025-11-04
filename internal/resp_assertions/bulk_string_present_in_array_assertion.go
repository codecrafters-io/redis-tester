package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type BulkStringPresentInArrayAssertion struct {
	ExpectedString string
}

func (a BulkStringPresentInArrayAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	array := value.Array()

	for _, element := range array {
		if element.Type == resp_value.BULK_STRING && element.String() == a.ExpectedString {
			return nil
		}
	}

	return fmt.Errorf("Expected bulk string '%s' to be present in the array, but is absent", a.ExpectedString)
}
