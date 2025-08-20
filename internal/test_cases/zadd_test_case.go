package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ZaddTestCase struct {
	Key                       string
	Member                    sorted_set.SortedSetMember
	ExpectedAddedMembersCount int
}

func (t *ZaddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	scoreStr := strconv.FormatFloat(t.Member.Score, 'f', -1, 64)
	sendCommandTestCase := SendCommandTestCase{
		Command:   "ZADD",
		Args:      []string{t.Key, scoreStr, t.Member.Name},
		Assertion: resp_assertions.NewIntegerAssertion(t.ExpectedAddedMembersCount),
	}
	return sendCommandTestCase.Run(client, logger)
}
