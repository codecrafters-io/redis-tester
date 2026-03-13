package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AuthTestCase struct {
	Username        string
	Password        string
	ExpectedSuccess bool
}

func (t AuthTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	var assertion resp_assertions.RESPAssertion

	if t.ExpectedSuccess {
		assertion = resp_assertions.NewSimpleStringAssertion("OK")
	} else {
		assertion = resp_assertions.PrefixAndSubstringsAssertion{
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        "WRONGPASS",
				CaseSensitive: true,
			},
		}
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "AUTH",
		Args:      []string{t.Username, t.Password},
		Assertion: assertion,
	}

	return sendCommandTestCase.Run(client, logger)
}
