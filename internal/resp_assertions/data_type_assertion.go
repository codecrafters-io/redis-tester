package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type DataTypeAssertion struct {
	ExpectedType string
}

func (a DataTypeAssertion) Run(value resp_value.Value) error {
	dataTypeHint := ""

	switch a.ExpectedType {
	case resp_value.NIL:
		dataTypeHint = "$-1\r\n"
	case resp_value.NIL_ARRAY:
		dataTypeHint = "*-1\r\n"
	}

	// Spacing
	if dataTypeHint != "" {
		dataTypeHint = " " + dataTypeHint
	}

	if value.Type != a.ExpectedType {
		return fmt.Errorf("Expected %s%s, found %s", a.ExpectedType, dataTypeHint, value.Type)
	}
	return nil
}
