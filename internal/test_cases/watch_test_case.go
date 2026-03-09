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
	// TODO: Will remove this after PR review
	// We'd wanna be lenient enough as we have been in case of errors previously. Eg.
	// See testPubSubSubscribe3 and others.
	// But, unlike other error cases, this error is worded as more 'english-like' so its hard to
	// expect a pattern. So, I left it like this.
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
