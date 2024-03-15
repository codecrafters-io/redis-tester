package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type WildcardCommandAssertion struct {
	ExpectedCommand string
	ExpectedArgs    []string
}

// NewWildcardCommandAssertion supports wildcards in command assertion.
// "*" asserts that a string is present, but we don't care about the value
// "**" asserts that a string is present, but we don't care about the value, and we don't care about the next elements
// "?" asserts that if a string is present, we will match the value
// capa eof capa psync2 => capa * ?capa *
func NewWildcardCommandAssertion(expectedCommand string, expectedArgs ...string) RESPAssertion {
	return WildcardCommandAssertion{
		ExpectedCommand: expectedCommand,
		ExpectedArgs:    expectedArgs,
	}
}

func (a WildcardCommandAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array type, got %s", value.Type)
	}

	elements := value.Array()

	if len(elements) < 1 {
		return fmt.Errorf("Expected array with at least 1 element, got %d elements", len(elements))
	}

	if elements[0].Type != resp_value.SIMPLE_STRING && elements[0].Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected first array element to be a string, got %s", elements[0].Type)
	}

	command := elements[0].String()

	if !strings.EqualFold(command, a.ExpectedCommand) {
		return fmt.Errorf("Expected command to be %q, got %q", strings.ToLower(a.ExpectedCommand), strings.ToLower(command))
	}

	for i, expectedArg := range a.ExpectedArgs {
		if expectedArg == "*" {
			actualArg := elements[i+1]
			// Don't compare actual value
			if actualArg.Type != resp_value.SIMPLE_STRING && actualArg.Type != resp_value.BULK_STRING {
				return fmt.Errorf("Expected argument %d to be a string, got %s", i+1, actualArg.Type)
			}
		} else if expectedArg[0] == '?' {
			// If actual argument is missing, continue
			if len(elements) > i+1 {
				actualArg := elements[i+1]
				if actualArg.Type != resp_value.SIMPLE_STRING && actualArg.Type != resp_value.BULK_STRING {
					return fmt.Errorf("Expected argument %d to be a string, got %s", i+1, actualArg.Type)
				}
				if actualArg.String() != expectedArg[1:] {
					return fmt.Errorf("Expected argument #%d to be %q, got %q", i+1, expectedArg, actualArg.String())
				}
			}
		} else if expectedArg == "**" {
			break
		} else {
			// Normal assertion
			actualArg := elements[i+1]
			if actualArg.Type != resp_value.SIMPLE_STRING && actualArg.Type != resp_value.BULK_STRING {
				return fmt.Errorf("Expected argument %d to be a string, got %s", i+1, actualArg.Type)
			}
			if actualArg.String() != expectedArg {
				return fmt.Errorf("Expected argument #%d to be %q, got %q", i+1, expectedArg, actualArg.String())
			}
		}
	}
	return nil
}
