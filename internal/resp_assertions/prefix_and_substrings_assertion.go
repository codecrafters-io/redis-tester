package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/logger"
)

// PrefixAndSubstringsAssertion should be used where we expect the value to either
// 1. Begins with a certain string
// 2. Contain one/multiple specified substrings
// 3. Both
// For complex patterns, use RegexAssertion
type PrefixAndSubstringsAssertion struct {
	ExpectedType           string
	PrefixPredicate        *PrefixPredicate
	Logger                 *logger.Logger
	HasSubstringPredicates []HasSubstringPredicate
}

func (a PrefixAndSubstringsAssertion) Run(value resp_value.Value) error {
	// This assertion is only valid for data types which is storable as byte slice
	if !resp_value.IsValueOfDataTypeStoredAsByteSlice(a.ExpectedType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - PrefixAndSubstringsAssertion is not applicable for %s", a.ExpectedType))
	}

	if a.Logger == nil {
		panic("Codecrafters Internal Error - Logger must be specified on PrefixAndSubstringsAssertion")
	}

	respErrorTypeAssertion := DataTypeAssertion{ExpectedType: a.ExpectedType}

	if err := respErrorTypeAssertion.Run(value); err != nil {
		return err
	}

	valueString := value.String()

	// Check the prefix pattern
	if a.PrefixPredicate != nil && !a.PrefixPredicate.Check(valueString) {
		hasTrailingSpace := ""

		if strings.HasSuffix(a.PrefixPredicate.Prefix, " ") {
			hasTrailingSpace = " (trailing space)"
		}

		return fmt.Errorf(
			"Expected %s to begin with %q%s, got %q",
			value.Type,
			a.PrefixPredicate.Prefix,
			hasTrailingSpace,
			valueString,
		)
	}

	substringsPresent := []string{}

	// Check for the specified substrings
	for _, hasSubstringCondition := range a.HasSubstringPredicates {
		if !hasSubstringCondition.Check(valueString) {
			// Print all the substrings that are present
			for _, substring := range substringsPresent {
				a.Logger.Infof("✔︎ Expected %s contains %q", value.Type, substring)
			}

			// Return error
			return fmt.Errorf("Expected %s to contain %q, got %q", value.Type, hasSubstringCondition.Substring, valueString)
		}
		substringsPresent = append(substringsPresent, hasSubstringCondition.Substring)
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
