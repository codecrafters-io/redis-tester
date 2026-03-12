package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type PatternedBytesAssertion struct {
	ExpectedType        string
	PrefixCondition     *PatternedBytesBeginsWithCondition
	SubstringConditions []PatternedBytesContainsCondition
}

func (a PatternedBytesAssertion) Run(value resp_value.Value) error {
	// This assertion is only valid for data types which is storable as byte slice
	if !resp_value.IsValueOfDataTypeStoredAsByteSlice(a.ExpectedType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - PatternedStringAssertion is not applicable for %s", a.ExpectedType))
	}

	respErrorTypeAssertion := DataTypeAssertion{ExpectedType: a.ExpectedType}

	if err := respErrorTypeAssertion.Run(value); err != nil {
		return err
	}

	valueString := value.String()

	// Check the prefix pattern
	if a.PrefixCondition != nil {
		if !a.PrefixCondition.Check(valueString) {
			return fmt.Errorf("Expected %s to begin with %q, got %q", value.Type, a.PrefixCondition.Prefix, valueString)
		}
	}

	for _, hasSubstringCondition := range a.SubstringConditions {
		if !hasSubstringCondition.Check(valueString) {
			return fmt.Errorf("Expected %s to contain %q, got %q", value.Type, hasSubstringCondition.Substring, valueString)
		}
	}

	return nil
}

type PatternedBytesBeginsWithCondition struct {
	Prefix        string
	CaseSensitive bool
}

func (c PatternedBytesBeginsWithCondition) Check(value string) bool {
	outerString := value
	innerString := c.Prefix

	if !c.CaseSensitive {
		outerString = strings.ToLower(outerString)
		innerString = strings.ToLower(innerString)
	}

	return strings.HasPrefix(outerString, innerString)
}

type PatternedBytesContainsCondition struct {
	Substring     string
	CaseSensitive bool
}

func (c PatternedBytesContainsCondition) Check(value string) bool {
	outerString := value
	innerString := c.Substring

	if !c.CaseSensitive {
		outerString = strings.ToLower(outerString)
		innerString = strings.ToLower(innerString)
	}

	return strings.Contains(outerString, innerString)
}
