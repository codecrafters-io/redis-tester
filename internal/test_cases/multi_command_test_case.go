package test_cases

import (
	"fmt"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// MultiCommandTestCase is a concise & easier way to define & run multiple SendCommandTestCase
type MultiCommandTestCase struct {
	Commands   [][]string
	Assertions []resp_assertions.RESPAssertion
}

func (t *MultiCommandTestCase) RunAll(client *resp_client.RespConnection, logger *logger.Logger) error {
	if len(t.Assertions) != len(t.Commands) {
		return fmt.Errorf("CodeCrafters internal error. Number of commands and assertions should be equal in MultiCommandTestCase")
	}

	for i, command := range t.Commands {
		setCommandTestCase := SendCommandTestCase{
			Command:   command[0],
			Args:      command[1:],
			Assertion: t.Assertions[i],
		}

		if err := setCommandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
