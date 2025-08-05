package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZadd2(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	zsetKey := testerutils_random.RandomWord()
	zsetSize := testerutils_random.RandomInt(4, 8)
	members := GenerateRandomZSetMembers(ZsetMemberGenerationOption{
		Count: zsetSize,
	})

	zsetTestCase := test_cases.NewZsetTestCase(zsetKey)
	for _, m := range members {
		zsetTestCase.AddMember(m.Name, m.Score)
	}

	// add all members
	if err := zsetTestCase.RunZaddAll(client, logger); err != nil {
		return err
	}

	// test by updating an existing element
	updateAtIdx := testerutils_random.RandomInt(0, zsetSize)
	newScore := strconv.FormatFloat(GetRandomZSetScore(), 'f', -1, 64)
	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "ZADD",
		Args:      []string{zsetKey, newScore, members[updateAtIdx].Name},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	return sendCommandTestCase.Run(client, logger)
}
