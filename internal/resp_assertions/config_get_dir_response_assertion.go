package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ConfigGetDirResponseAssertion struct {
	ExpectedDirValue string
}

func NewConfigGetDirResponseAssertion(expectedDirValue string) RESPAssertion {
	return ConfigGetDirResponseAssertion{ExpectedDirValue: expectedDirValue}
}

func (a ConfigGetDirResponseAssertion) Run(value resp_value.Value) error {
	arrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := arrayTypeAssertion.Run(value); err != nil {
		return err
	}

	if len(value.Array()) != 2 {
		return fmt.Errorf("Expected 2 elements in array, got %d (%s)", len(value.Array()), value.FormattedString())
	}

	firstElement := value.Array()[0]
	secondElement := value.Array()[1]

	err := NewBulkStringAssertion("dir").Run(firstElement)
	if err != nil {
		return err
	}

	if secondElement.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected element 1 to be a bulk string, got %s", secondElement.Type)
	}

	if secondElement.String() != a.ExpectedDirValue && secondElement.String() != a.ExpectedDirValue+"/" {
		return fmt.Errorf("Expected element 1 to be %q, got %q", a.ExpectedDirValue, secondElement.String())
	}

	return nil
}
