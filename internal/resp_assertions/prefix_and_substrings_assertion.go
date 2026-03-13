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
	Logger                 *logger.Logger
	HasPrefixPredicate     *PrefixPredicate
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

	dataTypeAssertion := DataTypeAssertion{ExpectedType: a.ExpectedType}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	valueString := value.String()

	// Check the prefix pattern
	if a.HasPrefixPredicate != nil && !a.HasPrefixPredicate.Check(valueString) {
		prefixDescription := ""

		// If the prefix has a trailing space but the value begins with the
		// prefix without the trailing space, notify
		if before, ok := strings.CutSuffix(a.HasPrefixPredicate.Prefix, " "); ok {
			prefixWithoutTrailingSpace := before

			prefixPredicateWithoutTrailingSpace := PrefixPredicate{
				Prefix:        prefixWithoutTrailingSpace,
				CaseSensitive: a.HasPrefixPredicate.CaseSensitive,
			}

			if prefixPredicateWithoutTrailingSpace.Check(valueString) {
				prefixDescription = " (trailing space)"
			}
		}

		return fmt.Errorf(
			"Expected %s to begin with %q%s, got %q",
			value.Type,
			a.HasPrefixPredicate.Prefix,
			prefixDescription,
			valueString,
		)
	}

	presentSubstrings := []string{}
	hasMissingSubstring := false
	firstMissingSubstring := ""

	// Check for the specified substrings
	for _, hasSubstringPredicate := range a.HasSubstringPredicates {
		if !hasSubstringPredicate.Check(valueString) {
			if !hasMissingSubstring {
				hasMissingSubstring = true
				firstMissingSubstring = hasSubstringPredicate.Substring
			}
		} else {
			presentSubstrings = append(presentSubstrings, hasSubstringPredicate.Substring)
		}
	}

	if !hasMissingSubstring {
		return nil
	}

	// Print all the present substrings first
	for _, presentSubstring := range presentSubstrings {
		a.Logger.Infof("✔︎ Expected %s contains %q", value.Type, presentSubstring)
	}

	return fmt.Errorf("Expected %s to contain %q, got %q", value.Type, firstMissingSubstring, valueString)
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
