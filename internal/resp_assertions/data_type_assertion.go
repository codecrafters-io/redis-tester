package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type DataTypeAssertion struct {
	ExpectedType string
}

func (a DataTypeAssertion) Run(value resp_value.Value) error {
	if value.Type != a.ExpectedType {
		return fmt.Errorf("Expected %s, found %s", a.ExpectedType, value.Type)
	}
	return nil
}
