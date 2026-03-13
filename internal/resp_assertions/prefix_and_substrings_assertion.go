package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// PrefixAndSubstringsAssertion should be used where we expect the value to either
// 1. Begins with a certain string
// 2. Contain one/multiple specified substrings
// 3. Both
// For complex patterns, use RegexAssertion
type PrefixAndSubstringsAssertion struct {
	ExpectedType           string
	PrefixPredicate        *PrefixPredicate
	HasSubstringPredicates []HasSubstringPredicate
}

func (a PrefixAndSubstringsAssertion) Run(value resp_value.Value) error {
	// This assertion is only valid for data types which is storable as byte slice
	if !resp_value.IsValueOfDataTypeStoredAsByteSlice(a.ExpectedType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - PrefixAndSubstringsAssertion is not applicable for %s", a.ExpectedType))
	}

	respErrorTypeAssertion := DataTypeAssertion{ExpectedType: a.ExpectedType}

	if err := respErrorTypeAssertion.Run(value); err != nil {
		return err
	}

	valueString := value.String()

	// Check the prefix pattern
	if a.PrefixPredicate != nil {
		if !a.PrefixPredicate.Check(valueString) {
			return fmt.Errorf("Expected %s to begin with %q, got %q", value.Type, a.PrefixPredicate.Prefix, valueString)
		}
	}

	// Check for the specified substrings
	for _, hasSubstringCondition := range a.HasSubstringPredicates {
		if !hasSubstringCondition.Check(valueString) {
			return fmt.Errorf("Expected %s to contain %q, got %q", value.Type, hasSubstringCondition.Substring, valueString)
		}
	}

	return nil
}

type PrefixPredicate struct {
	Prefix        string
	CaseSensitive bool
}

func (c PrefixPredicate) Check(value string) bool {
	outerString := value
	innerString := c.Prefix

	if !c.CaseSensitive {
		outerString = strings.ToLower(outerString)
		innerString = strings.ToLower(innerString)
	}

	return strings.HasPrefix(outerString, innerString)
}

type HasSubstringPredicate struct {
	Substring     string
	CaseSensitive bool
}

func (c HasSubstringPredicate) Check(value string) bool {
	outerString := value
	innerString := c.Substring

	if !c.CaseSensitive {
		outerString = strings.ToLower(outerString)
		innerString = strings.ToLower(innerString)
	}

	return strings.Contains(outerString, innerString)
}
