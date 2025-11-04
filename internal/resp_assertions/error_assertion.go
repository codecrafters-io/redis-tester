package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ErrorAssertion struct {
	ExpectedValue string
}

func NewErrorAssertion(expectedValue string) RESPAssertion {
	return ErrorAssertion{ExpectedValue: expectedValue}
}

func (a ErrorAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ERROR}
	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	if value.Error() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.Error())
	}

	return nil
}
