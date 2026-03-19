package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// ConfigGetBulkStringValueAssertion should be used when a command is to be issued 'CONFIG GET key' and the
// response is of the form ["key", "value"], encoded as array of two bulk strings
type ConfigGetBulkStringValueAssertion struct {
	ExpectedKey   string
	ExpectedValue string
}

func NewConfigGetBulkStringValueAssertion(expectedKey string, expectedValue string) RESPAssertion {
	return ConfigGetBulkStringValueAssertion{
		ExpectedKey:   expectedKey,
		ExpectedValue: expectedValue,
	}
}

func (a ConfigGetBulkStringValueAssertion) Run(value resp_value.Value) error {
	arrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := arrayTypeAssertion.Run(value); err != nil {
		return err
	}

	if len(value.Array()) != 2 {
		return fmt.Errorf("Expected 2 elements in array, got %d (%s)", len(value.Array()), value.FormattedString())
	}

	firstElement := value.Array()[0]
	secondElement := value.Array()[1]

	if firstElement.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected element #1 to be a bulk string, got %s", firstElement.Type)
	}

	if firstElement.String() != a.ExpectedKey {
		return fmt.Errorf("Expected element #1 to be %q, got %q", a.ExpectedKey, firstElement.String())
	}

	if secondElement.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected element #2 to be a bulk string, got %s", secondElement.Type)
	}
	if secondElement.String() != a.ExpectedValue && secondElement.String() != a.ExpectedValue+"/" {
		return fmt.Errorf("Expected element #2 to be %q, got %q", a.ExpectedValue, secondElement.String())
	}

	return nil
}
