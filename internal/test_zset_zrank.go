package internal

import (
	"fmt"

	ds "github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZrank(stageHarness *test_case_harness.TestCaseHarness) error {
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
	sortedSet := ds.GenerateZsetWithRandomMembers(ds.ZsetMemberGenerationOption{
		Count:          testerutils_random.RandomInt(4, 8),
		SameScoreCount: 2,
	})
	members := sortedSet.GetMembers()

	shuffledMembers := testerutils_random.ShuffleArray(members)
	for _, m := range shuffledMembers {
		zaddTestCase := test_cases.ZaddTestCase{
			Key:                  zsetKey,
			Member:               m,
			ExpectedAddedMembers: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Run zrank for random elements
	ranksToTest := testerutils_random.RandomInts(0, sortedSet.Size(), sortedSet.Size()/2)
	for _, rank := range ranksToTest {
		zrankTestCase := test_cases.SendCommandTestCase{
			Command:   "ZRANK",
			Args:      []string{zsetKey, members[rank].GetName()},
			Assertion: resp_assertions.NewIntegerAssertion(rank),
		}

		if err := zrankTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Test ranks using missing key and missing member
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))
	missingMember := fmt.Sprintf("missing_member_%d", testerutils_random.RandomInt(1, 100))
	missingValuesZrankTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"ZRANK", zsetKey, missingMember},
				Assertion: resp_assertions.NewNilAssertion(),
			},
			{
				Command:   []string{"ZRANK", missingKey, members[0].GetName()},
				Assertion: resp_assertions.NewNilAssertion(),
			},
		},
	}

	return missingValuesZrankTestCase.RunAll(client, logger)
}
