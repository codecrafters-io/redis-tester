package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZrangePosIndex(stageHarness *test_case_harness.TestCaseHarness) error {
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

	// Add members
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

	memberNames := sortedSet.GetMemberNames()
	middleIndex := testerutils_random.RandomInt(1, sortedSet.Size()-1)
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	zrangeTestCases := []test_cases.ZrangeTestCase{
		// usual test cases
		{
			Key:                 zsetKey,
			StartIndex:          0,
			EndIndex:            middleIndex,
			ExpectedMemberNames: memberNames[0 : middleIndex+1],
		},
		{
			Key:                 zsetKey,
			StartIndex:          middleIndex,
			EndIndex:            sortedSet.Size() - 1,
			ExpectedMemberNames: memberNames[middleIndex:],
		},
		{
			Key:                 zsetKey,
			StartIndex:          0,
			EndIndex:            sortedSet.Size() - 1,
			ExpectedMemberNames: memberNames,
		},
		// start Index > end index
		{
			Key:                 zsetKey,
			StartIndex:          1,
			EndIndex:            0,
			ExpectedMemberNames: []string{},
		},
		// end index out of bounds
		{
			Key:                 zsetKey,
			StartIndex:          0,
			EndIndex:            sortedSet.Size() * 2,
			ExpectedMemberNames: memberNames,
		},
		// key does not exist
		{
			Key:                 missingKey,
			StartIndex:          0,
			EndIndex:            1,
			ExpectedMemberNames: []string{},
		},
	}

	for _, zrangeTestCase := range zrangeTestCases {
		if err := zrangeTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
