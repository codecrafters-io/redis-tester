package resp_assertions

import (
	"encoding/json"
	"fmt"
	"reflect"

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

func (a XReadResponseAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	expected := a.normalizeExpected()
	actual := normalizeActual(value)

	if !reflect.DeepEqual(expected, actual) {
		expectedJSON, err := json.MarshalIndent(expected, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal expected value: %w", err)
		}
		actualJSON, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal actual value: %w", err)
		}
		return fmt.Errorf("XREAD response mismatch:\nExpected:\n%s\nGot:\n%s", expectedJSON, actualJSON)
	}

	return nil
}

func (a XReadResponseAssertion) normalizeExpected() []interface{} {
	result := make([]interface{}, len(a.ExpectedStreamResponses))

	for i, stream := range a.ExpectedStreamResponses {
		entries := make([][]interface{}, 0, len(stream.Entries))
		for _, entry := range stream.Entries {
			flatPairs := make([]string, 0, len(entry.FieldValuePairs)*2)
			for _, pair := range entry.FieldValuePairs {
				flatPairs = append(flatPairs, pair[0], pair[1])
			}
			entries = append(entries, []interface{}{entry.Id, flatPairs})
		}
		result[i] = []interface{}{stream.Key, entries}
	}

	return result
}

func normalizeActual(v resp_value.Value) interface{} {
	switch v.Type {
	case resp_value.BULK_STRING:
		return v.String()
	case resp_value.ARRAY:
		arr := v.Array()
		result := make([]interface{}, len(arr))
		for i, elem := range arr {
			// Special handling for stream entries (second element of the outer array)
			if i == 1 && len(arr) == 2 {
				entries := elem.Array()
				typedEntries := make([][]interface{}, len(entries))
				for j, entry := range entries {
					entryArr := entry.Array()
					if len(entryArr) == 2 {
						fvPairs := entryArr[1].Array()
						strPairs := make([]string, len(fvPairs))
						for k, pair := range fvPairs {
							strPairs[k] = pair.String()
						}
						typedEntries[j] = []interface{}{normalizeActual(entryArr[0]), strPairs}
					}
				}
				result[i] = typedEntries
				continue
			}
			result[i] = normalizeActual(elem)
		}
		return result
	default:
		return v.String()
	}
}
