package resp_assertions

import (
	"errors"
	"fmt"
	"regexp"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

// RegexAssertion should be used if the expected value must match a given
// regex pattern. For simple use cases, use PatternedBytesAssertion
type RegexAssertion struct {
	ExpectedType    string
	ExpectedPattern *regexp.Regexp
	ExpectedSample  string
}

func NewRegexAssertion(expectedType string, expectedPattern string, expectedSample string) RegexAssertion {
	return RegexAssertion{
		ExpectedType:    expectedType,
		ExpectedPattern: regexp.MustCompile(expectedPattern),
		ExpectedSample:  expectedSample,
	}
}

func (a RegexAssertion) Run(value resp_value.Value) error {
	// This assertion is only valid for data types which is storable as byte slice
	if !resp_value.IsValueOfDataTypeStoredAsByteSlice(a.ExpectedType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - RegexAssertion is not applicable for %s", a.ExpectedType))
	}

	// Ensure that the expected sample matches the regex
	if !a.ExpectedPattern.Match([]byte(a.ExpectedSample)) {
		panic("Codecrafters Internal Error - RegexAssertion: ExpectedSample does not match ExpectedPattern")
	}

	dataTypeAssertion := DataTypeAssertion{ExpectedType: a.ExpectedType}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	receivedValueString := value.String()

	// Match against the regex pattern
	if !a.ExpectedPattern.Match([]byte(receivedValueString)) {
		return errors.New(BuildExpectedVsReceivedErrorMessage(a.ExpectedSample, receivedValueString))
	}

	return nil
}
