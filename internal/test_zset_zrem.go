package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZrem(stageHarness *test_case_harness.TestCaseHarness) error {
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

	if err := zsetTestCase.RunZaddAll(client, logger); err != nil {
		return err
	}

	idxToRemove := testerutils_random.RandomInt(0, zsetSize)
	memberToRemove := members[idxToRemove].Name
	zremTestCase := test_cases.SendCommandTestCase{
		Command:   "ZREM",
		Args:      []string{zsetKey, memberToRemove},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}
	if err := zremTestCase.Run(client, logger); err != nil {
		return err
	}

	zsetTestCase.RemoveMember(memberToRemove)
	if err := zsetTestCase.RunZrange(client, logger, 0, -1); err != nil {
		return err
	}

	/* Remove a missing member */
	missing_member := fmt.Sprintf("missing_member_%d", testerutils_random.RandomInt(1, 100))
	zremTestCase = test_cases.SendCommandTestCase{
		Command:   "ZREM",
		Args:      []string{zsetKey, missing_member},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	if err := zremTestCase.Run(client, logger); err != nil {
		return err
	}

	/* Zset shouldn't change */
	return zsetTestCase.RunZrange(client, logger, 0, -1)
}
