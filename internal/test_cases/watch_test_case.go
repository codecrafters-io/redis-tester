package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type WatchTestCase struct {
	Keys                []string
	IsInsideTransaction bool
}

func (t WatchTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	assertion := resp_assertions.NewSimpleStringAssertion("OK")

	// Must assert error inside a transaction
	if t.IsInsideTransaction {
		assertion = resp_assertions.PrefixAndSubstringsAssertion{
			Logger:       logger,
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        "ERR ",
				CaseSensitive: true,
			},
			HasSubstringPredicates: []resp_assertions.HasSubstringPredicate{
				{
					Substring: "watch",
				},
				{
					Substring: "inside multi",
				},
				{
					Substring: "not allowed",
				},
			},
		}
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "WATCH",
		Args:      t.Keys,
		Assertion: assertion,
	}

	return sendCommandTestCase.Run(client, logger)
}
