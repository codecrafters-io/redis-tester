package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AclWhoamiErrorTestCase struct {
	ExpectedErrorBeginsWith string
}

func (t AclWhoamiErrorTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command: "ACL",
		Args:    []string{"WHOAMI"},
		Assertion: resp_assertions.PrefixAndSubstringsAssertion{
			Logger:       logger,
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        t.ExpectedErrorBeginsWith,
				CaseSensitive: true,
			},
		},
	}

	return sendCommandTestCase.Run(client, logger)
}
