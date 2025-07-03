package test_cases

import (
	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type CommandWithAssertion struct {
	Command   []string
	Assertion resp_assertions.RESPAssertion
}

// MultiCommandTestCase is a concise & easier way to define & run multiple SendCommandTestCase
type MultiCommandTestCase struct {
	CommandWithAssertions []CommandWithAssertion
}

func (t *MultiCommandTestCase) RunAll(client *resp_client.RespConnection, logger *logger.Logger) error {
	for _, cwa := range t.CommandWithAssertions {
		setCommandTestCase := SendCommandTestCase{
			Command:   cwa.Command[0],
			Args:      cwa.Command[1:],
			Assertion: cwa.Assertion,
		}

		if err := setCommandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
