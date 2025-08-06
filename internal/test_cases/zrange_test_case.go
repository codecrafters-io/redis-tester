package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ZrangeTestCase struct {
	Key                 string
	StartIndex          int
	EndIndex            int
	ExpectedMemberNames []string
}

func (t *ZrangeTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	startIdxStr := strconv.Itoa(t.StartIndex)
	endIdxStr := strconv.Itoa(t.EndIndex)
	sendCommandTestCase := SendCommandTestCase{
		Command:   "ZRANGE",
		Args:      []string{t.Key, startIdxStr, endIdxStr},
		Assertion: resp_assertions.NewOrderedStringArrayAssertion(t.ExpectedMemberNames),
	}

	return sendCommandTestCase.Run(client, logger)
}
