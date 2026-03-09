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

	if t.IsInsideTransaction {
		assertion = resp_assertions.NewErrorAssertion("ERR WATCH inside MULTI is not allowed")
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "WATCH",
		Args:      t.Keys,
		Assertion: assertion,
	}

	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
