package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
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
		assertion = resp_assertions.NewRegexErrorAssertion("(?i)not allowed")
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "WATCH",
		Args:      t.Keys,
		Assertion: assertion,
	}

	return sendCommandTestCase.Run(client, logger)
}
