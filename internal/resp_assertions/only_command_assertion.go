package resp_assertions

import (
	"fmt"
	"strings"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type OnlyCommandAssertion struct {
	ExpectedCommand string
}

func NewOnlyCommandAssertion(expectedCommand string) RESPAssertion {
	return OnlyCommandAssertion{
		ExpectedCommand: expectedCommand,
	}
}

func (a OnlyCommandAssertion) Run(value resp_value.Value) error {
	arrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := arrayTypeAssertion.Run(value); err != nil {
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

	return nil
}
