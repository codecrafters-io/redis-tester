package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
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
	sortedSet := sorted_set.GenerateSortedSetWithRandomMembers(sorted_set.SortedSetMemberGenerationOption{
		Count: testerutils_random.RandomInt(2, 4),
	})
	members := sortedSet.GetMembers()

	// Test using new members
	for _, m := range members {
		zaddTestCase := test_cases.ZaddTestCase{
			Key:                       zsetKey,
			Member:                    m,
			ExpectedAddedMembersCount: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Test by updating an existing member
	memberToUpdate := members[testerutils_random.RandomInt(0, sortedSet.Size())]
	newScore := sorted_set.GetRandomSortedSetScore()
	zaddTestCase := test_cases.ZaddTestCase{
		Key: zsetKey,
		Member: sorted_set.SortedSetMember{
			Name:  memberToUpdate.Name,
			Score: newScore,
		},
		ExpectedAddedMembersCount: 0,
	}

	return zaddTestCase.Run(client, logger)
}
