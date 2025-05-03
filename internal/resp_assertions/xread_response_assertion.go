package resp_assertions

import (
	"encoding/json"
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type StreamEntry struct {
	Id              string
	FieldValuePairs [][]string
}

type StreamResponse struct {
	Key     string
	Entries []StreamEntry
}

type XReadResponseAssertion struct {
	ExpectedStreamResponses []StreamResponse
}

func NewXReadResponseAssertion(expectedValue []StreamResponse) RESPAssertion {
	return XReadResponseAssertion{ExpectedStreamResponses: expectedValue}
}

func (a XReadResponseAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	expected := a.buildExpected().ToSerializable()
	actual := value.ToSerializable()

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal expected value: %w", err)
	}
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal actual value: %w", err)
	}

	if string(expectedJSON) != string(actualJSON) {
		return fmt.Errorf("XREAD response mismatch:\nExpected:\n%s\nGot:\n%s", expectedJSON, actualJSON)
	}

	return nil
}

func (a XReadResponseAssertion) buildExpected() resp_value.Value {
	streams := make([]resp_value.Value, len(a.ExpectedStreamResponses))
	for i, stream := range a.ExpectedStreamResponses {
		streams[i] = stream.toRESPValue()
	}
	return resp_value.NewArrayValue(streams)
}

func (s StreamResponse) toRESPValue() resp_value.Value {
	entries := make([]resp_value.Value, len(s.Entries))
	for i, entry := range s.Entries {
		entries[i] = entry.toRESPValue()
	}
	return resp_value.NewArrayValue([]resp_value.Value{
		resp_value.NewBulkStringValue(s.Key),
		resp_value.NewArrayValue(entries),
	})
}

func (e StreamEntry) toRESPValue() resp_value.Value {
	var fieldValues []resp_value.Value
	for _, pair := range e.FieldValuePairs {
		for _, v := range pair {
			fieldValues = append(fieldValues, resp_value.NewBulkStringValue(v))
		}
	}
	return resp_value.NewArrayValue([]resp_value.Value{
		resp_value.NewBulkStringValue(e.Id),
		resp_value.NewArrayValue(fieldValues),
	})
}
