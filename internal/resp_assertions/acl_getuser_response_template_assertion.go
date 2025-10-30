package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type AclGetUserResponseTemplateAssertion struct {
	AssertForFlags    bool
	AsserForPasswords bool
}

func (a AclGetUserResponseTemplateAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	array := value.Array()

	if a.AssertForFlags {
		if err := a.assertForFlags(array); err != nil {
			return err
		}
	}

	if a.AsserForPasswords {
		if err := a.assertForPasswords(array); err != nil {
			return err
		}
	}

	return nil
}

func (a AclGetUserResponseTemplateAssertion) assertForFlags(array []resp_value.Value) error {
	// Expect length of array to be at least 2 in case of flags
	if len(array) < 2 {
		return fmt.Errorf("Expected the array length to be at least 2 for flags to be present, got %d", len(array))
	}

	// Assert the first element to be "flags"
	firstElement := array[0]
	flagsLiteralAssertion := NewStringAssertion("flags")

	if err := flagsLiteralAssertion.Run(firstElement); err != nil {
		return fmt.Errorf("Expected the first element of the array to be \"flags\", got %s", firstElement.FormattedString())
	}

	// Assert type of the second element (array)
	secondElement := array[1]

	if secondElement.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected the second element to be of type array, got %s", secondElement.Type)
	}

	for i, value := range secondElement.Array() {
		if value.Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected the second element of the array to have only bulk strings, got %s at index %d", value.Type, i)
		}
	}

	return nil
}

func (a AclGetUserResponseTemplateAssertion) assertForPasswords(array []resp_value.Value) error {
	// Expect length of array to be at least 4 in case of passwords
	if len(array) < 4 {
		return fmt.Errorf("Expected the array length to be at least 4 for passwords to be present, got %d", len(array))
	}

	thirdElement := array[2]
	passwordsLiteralAssertion := NewStringAssertion("passwords")

	if err := passwordsLiteralAssertion.Run(thirdElement); err != nil {
		return fmt.Errorf("Expected the third element of the array to be \"passwords\", got %s", thirdElement.FormattedString())
	}

	// Assert type of fourth element (array)
	fourthElement := array[3]

	if array[3].Type != resp_value.ARRAY {
		return fmt.Errorf("Expected the fourth element to be of type array, got %s", fourthElement.Type)
	}

	for i, value := range fourthElement.Array() {
		if value.Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected the fourth element of the array to have only bulk strings, got %s at index %d", value.Type, i)
		}
	}

	return nil
}
