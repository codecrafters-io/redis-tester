package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AclWhoamiTestCase struct {
	ExpectedUsername string
}

func (t AclWhoamiTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"WHOAMI"},
		Assertion: resp_assertions.NewBulkStringAssertion(t.ExpectedUsername),
	}

	return sendCommandTestCase.Run(client, logger)
}
