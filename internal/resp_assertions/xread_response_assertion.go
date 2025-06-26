package resp_assertions

import (
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

	expectedValue := a.buildExpected()
	expected := expectedValue.FormattedString()

	actual := value.FormattedString()

	if expected != actual {
		return fmt.Errorf("XREAD response mismatch:\nExpected:\n%s\nGot:\n%s", expected, actual)
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
