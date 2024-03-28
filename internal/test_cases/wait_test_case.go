package test_cases

import (
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type WaitTestCase struct {
	Replicas        string
	Timeout         string
	ExpectedMessage int
}

func (t WaitTestCase) RunWait(client *resp_connection.RespConnection, logger *logger.Logger) error {
	commandTest := SendCommandTestCase{
		Command:   "WAIT",
		Args:      []string{t.Replicas, t.Timeout},
		Assertion: resp_assertions.NewIntegerAssertion(t.ExpectedMessage),
	}

	return commandTest.Run(client, logger)
}
