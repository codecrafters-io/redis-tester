package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type PublishTestCase struct {
	Channel                 string
	Message                 string
	ExpectedSubscriberCount int
}

func (t *PublishTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command:   "PUBLISH",
		Args:      []string{t.Channel, t.Message},
		Assertion: resp_assertions.NewIntegerAssertion(t.ExpectedSubscriberCount),
	}

	return sendCommandTestCase.Run(client, logger)
}
