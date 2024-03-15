package test_cases

import (
	"strconv"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ReplicationTestCase struct{}

func (t ReplicationTestCase) RunGetAck(client *resp_connection.RespConnection, logger *logger.Logger, offset int) error {
	commandTest := SendCommandAndReceiveValueTestCase{
		Command:   "REPLCONF",
		Args:      []string{"GETACK", "*"},
		Assertion: resp_assertions.NewCommandAssertion("REPLCONF", "ACK", strconv.Itoa(offset)),
	}

	return commandTest.Run(client, logger)
}

func (t ReplicationTestCase) RunWait(client *resp_connection.RespConnection, logger *logger.Logger, replicas string, timeout string, expectedMessage int) error {
	commandTest := SendCommandAndReceiveValueTestCase{
		Command:   "WAIT",
		Args:      []string{replicas, timeout},
		Assertion: resp_assertions.NewIntegerAssertion(expectedMessage),
	}

	return commandTest.Run(client, logger)
}
