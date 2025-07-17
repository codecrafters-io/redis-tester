package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type UnsubscribeTestCase struct {
	Channel                                 string
	ExpectedSubscriberCountAfterUnsubscribe int
}

func (t *UnsubscribeTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command: "UNSUBSCRIBE",
		Args:    []string{t.Channel},
		Assertion: resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("unsubscribe"),
			resp_assertions.NewStringAssertion(t.Channel),
			resp_assertions.NewIntegerAssertion(t.ExpectedSubscriberCountAfterUnsubscribe),
		}),
	}

	return sendCommandTestCase.Run(client, logger)
}
