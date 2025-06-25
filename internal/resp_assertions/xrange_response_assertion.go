package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type XRangeResponseAssertion struct {
	ExpectedStreamResponses []StreamEntry
}

func NewXRangeResponseAssertion(expected []StreamEntry) RESPAssertion {
	return XRangeResponseAssertion{
		ExpectedStreamResponses: expected,
	}
}

func (x XRangeResponseAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	expectedValue := x.buildExpected()
	expected := expectedValue.FormattedString()
	actual := value.FormattedString()

	if expected != actual {
		return fmt.Errorf("XRANGE response mismatch:\nExpected:\n%s\nGot:\n%s", expected, actual)
	}

	return nil
}

func (x XRangeResponseAssertion) buildExpected() resp_value.Value {
	entries := make([]resp_value.Value, len(x.ExpectedStreamResponses))
	for i, entry := range x.ExpectedStreamResponses {
		entries[i] = entry.toRESPValue()
	}
	return resp_value.NewArrayValue(entries)
}
