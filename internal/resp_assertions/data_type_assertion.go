package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type DataTypeAssertion struct {
	ExpectedType string
}

func (a DataTypeAssertion) Run(value resp_value.Value) error {
	if value.Type == a.ExpectedType {
		return nil
	}

	expectedDataTypeHint := a.getDataTypeHint(a.ExpectedType)
	receivedDataTypeHint := a.getDataTypeHint(value.Type)

	// Spacing
	if expectedDataTypeHint != "" {
		expectedDataTypeHint = fmt.Sprintf(" (%s)", expectedDataTypeHint)
	}

	if receivedDataTypeHint != "" {
		receivedDataTypeHint = fmt.Sprintf(" (%s)", receivedDataTypeHint)
	}

	return fmt.Errorf(
		"Expected %s%s, found %s%s",
		a.ExpectedType, expectedDataTypeHint,
		value.Type, receivedDataTypeHint,
	)
}

func (a DataTypeAssertion) getDataTypeHint(dataType string) string {
	dataTypeHint := ""
	switch dataType {
	case resp_value.NIL:
		dataTypeHint = "$-1\r\n"
	case resp_value.NIL_ARRAY:
		dataTypeHint = "*-1\r\n"
	}

	if dataTypeHint != "" {
		dataTypeHint = FormatWithoutQuotes(dataTypeHint)
	}

	return dataTypeHint
}
