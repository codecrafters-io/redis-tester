package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GetAckTestCase struct{}

func (t GetAckTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger, offset int) error {
	commandTest := SendCommandTestCase{
		Command:   "REPLCONF",
		Args:      []string{"GETACK", "*"},
		Assertion: resp_assertions.NewCommandAssertion("REPLCONF", "ACK", strconv.Itoa(offset)),
	}

	return commandTest.Run(client, logger)
}
