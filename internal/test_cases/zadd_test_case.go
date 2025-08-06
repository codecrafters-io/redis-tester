package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ZaddTestCase struct {
	Key                  string
	Member               data_structures.SortedSetMember
	ExpectedAddedMembers int
}

func (t *ZaddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	scoreStr := strconv.FormatFloat(t.Member.GetScore(), 'f', -1, 64)
	sendCommandTestCase := SendCommandTestCase{
		Command:   "ZADD",
		Args:      []string{t.Key, scoreStr, t.Member.GetName()},
		Assertion: resp_assertions.NewIntegerAssertion(t.ExpectedAddedMembers),
	}
	return sendCommandTestCase.Run(client, logger)
}
