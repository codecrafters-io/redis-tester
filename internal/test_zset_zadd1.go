package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZadd1(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keyMemberPair := testerutils_random.RandomWords(2)

	zsetKey := keyMemberPair[0]
	sortedSet := sorted_set.GenerateSortedSetWithRandomMembers(sorted_set.SortedSetMemberGenerationOption{
		Count: 1,
	})
	member := sortedSet.GetMembers()[0]

	zaddTestCase := test_cases.ZaddTestCase{
		Key: zsetKey,
		Member: sorted_set.SortedSetMember{
			Name:  member.Name,
			Score: member.Score,
		},
		ExpectedAddedMembersCount: 1,
	}

	return zaddTestCase.Run(client, logger)
}
