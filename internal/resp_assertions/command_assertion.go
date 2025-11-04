package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type CommandAssertion struct {
	ExpectedCommand string
	ExpectedArgs    []string
}

func NewCommandAssertion(expectedCommand string, expectedArgs ...string) RESPAssertion {
	return CommandAssertion{
		ExpectedCommand: expectedCommand,
		ExpectedArgs:    expectedArgs,
	}
}

func (a CommandAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	elements := value.Array()

	if len(elements) < 1 {
		return fmt.Errorf("Expected array with at least 1 element, got %d elements", len(elements))
	}

	if elements[0].Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected first array element to be a bulk string, got %s", elements[0].Type)
	}

	command := elements[0].String()

	if !strings.EqualFold(command, a.ExpectedCommand) {
		return fmt.Errorf("Expected command to be %q, got %q", strings.ToLower(a.ExpectedCommand), strings.ToLower(command))
	}

	if len(elements) != len(a.ExpectedArgs)+1 {
		return fmt.Errorf("Expected command to have %d arguments, got %d", len(a.ExpectedArgs), len(elements)-1)
	}

	for i, expectedArg := range a.ExpectedArgs {
		actualArg := elements[i+1]
		if actualArg.Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected argument %d to be a bulk string, got %s", i+1, actualArg.Type)
		}

		if actualArg.String() != expectedArg {
			return fmt.Errorf("Expected argument #%d to be %q, got %q", i+1, expectedArg, actualArg.String())
		}
	}

	return nil
}
