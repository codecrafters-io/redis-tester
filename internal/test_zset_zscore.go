package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZscore(stageHarness *test_case_harness.TestCaseHarness) error {
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
		Count:          testerutils_random.RandomInt(3, 5),
		SameScoreCount: 2,
	})
	members := sortedSet.GetMembers()

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

	zscoreTolerance := 1e-10

	memberToTest := members[testerutils_random.RandomInt(0, sortedSet.Size())]
	zscoreTestCase := test_cases.SendCommandTestCase{
		Command:   "ZSCORE",
		Args:      []string{zsetKey, memberToTest.Name},
		Assertion: resp_assertions.NewFloatingPointBulkStringAssertion(memberToTest.Score, zscoreTolerance),
	}

	if err := zscoreTestCase.Run(client, logger); err != nil {
		return err
	}

	// Update an existing member
	newScore := sorted_set.GetRandomSortedSetScore()
	zaddTestCase := test_cases.ZaddTestCase{
		Key: zsetKey,
		Member: sorted_set.SortedSetMember{
			Name:  memberToTest.Name,
			Score: newScore,
		},
		ExpectedAddedMembersCount: 0,
	}

	if err := zaddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Test the score
	zscoreTestCase = test_cases.SendCommandTestCase{
		Command:   "ZSCORE",
		Args:      []string{zsetKey, memberToTest.Name},
		Assertion: resp_assertions.NewFloatingPointBulkStringAssertion(newScore, zscoreTolerance),
	}

	return zscoreTestCase.Run(client, logger)
}
