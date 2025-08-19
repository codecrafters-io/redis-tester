package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
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
	sortedSet := sorted_set.GenerateSortedSetWithRandomMembers(sorted_set.SortedSetMemberGenerationOption{
		Count: testerutils_random.RandomInt(2, 4),
	})
	members := sortedSet.GetMembers()

	// add members
	shuffledMembers := testerutils_random.ShuffleArray(members)
	for _, m := range shuffledMembers {
		zaddTestCase := test_cases.ZaddTestCase{
			Key:                       zsetKey,
			Member:                    m,
			ExpectedAddedMembersCount: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// remove a member
	memberToRemove := members[testerutils_random.RandomInt(0, sortedSet.Size())].Name
	zremTestCase := test_cases.SendCommandTestCase{
		Command:   "ZREM",
		Args:      []string{zsetKey, memberToRemove},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}
	if err := zremTestCase.Run(client, logger); err != nil {
		return err
	}

	// check remaining members
	sortedSet.RemoveMember(memberToRemove)
	updatedMemberNames := sortedSet.GetMemberNames()
	zrangeTestCase := test_cases.ZrangeTestCase{
		Key:                 zsetKey,
		StartIndex:          0,
		EndIndex:            -1,
		ExpectedMemberNames: updatedMemberNames,
	}
	if err := zrangeTestCase.Run(client, logger); err != nil {
		return err
	}

	// remove a missing member
	missing_member := fmt.Sprintf("missing_member_%d", testerutils_random.RandomInt(1, 100))
	zremTestCase = test_cases.SendCommandTestCase{
		Command:   "ZREM",
		Args:      []string{zsetKey, missing_member},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	if err := zremTestCase.Run(client, logger); err != nil {
		return err
	}

	// verify that the sorted set hasn't changed
	return zrangeTestCase.Run(client, logger)
}
