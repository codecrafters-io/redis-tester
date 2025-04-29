package resp_assertions

import (
	"encoding/json"
	"fmt"
	"reflect"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StreamAssertion struct {
	ExpectedValue [][]interface{}
}

func NewStreamAssertion(expectedValue [][]interface{}) RESPAssertion {
	return StreamAssertion{ExpectedValue: expectedValue}
}

func (a StreamAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	var convertToSlice func(v resp_value.Value) interface{}
	convertToSlice = func(v resp_value.Value) interface{} {
		switch v.Type {
		case resp_value.BULK_STRING:
			return v.String()
		case resp_value.ARRAY:
			result := make([]interface{}, len(v.Array()))
			for i, elem := range v.Array() {
				result[i] = convertToSlice(elem)
			}
			return result
		default:
			return v.String()
		}
	}
	actual := convertToSlice(value).([]interface{})

	expected := make([]interface{}, len(a.ExpectedValue))
	for i, v := range a.ExpectedValue {
		expected[i] = v
	}

	if !reflect.DeepEqual(expected, actual) {
		expectedJSON, err := json.MarshalIndent(expected, "", "  ")
		if err != nil {
			return err
		}
		actualJSON, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			return err
		}
		return fmt.Errorf("Expected:\n%s\nGot:\n%s", expectedJSON, actualJSON)
	}

	return nil
}
