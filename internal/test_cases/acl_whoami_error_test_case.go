package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AclWhoamiErrorTestCase struct {
	ExpectedErrorPattern string
}

func (t AclWhoamiErrorTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"WHOAMI"},
		Assertion: resp_assertions.NewRegexErrorAssertion(t.ExpectedErrorPattern),
	}

	return sendCommandTestCase.Run(client, logger)
}
