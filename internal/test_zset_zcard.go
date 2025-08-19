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

func testZsetZcard(stageHarness *test_case_harness.TestCaseHarness) error {
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
	for i, m := range shuffledMembers {
		// Add member
		zaddTestCase := test_cases.ZaddTestCase{
			Key:                       zsetKey,
			Member:                    m,
			ExpectedAddedMembersCount: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}

		// Test for cardinality
		zcardTestCase := test_cases.SendCommandTestCase{
			Command:   "ZCARD",
			Args:      []string{zsetKey},
			Assertion: resp_assertions.NewIntegerAssertion(i + 1),
		}
		if err := zcardTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Update an existing member
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
	if err := zaddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Check the cardinality again: size shouldn't change
	zcardTestCase := test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{zsetKey},
		Assertion: resp_assertions.NewIntegerAssertion(sortedSet.Size()),
	}
	if err := zcardTestCase.Run(client, logger); err != nil {
		return err
	}

	// Check ZCARD with a missing key
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))
	missingKeyZcardTestCase := test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{missingKey},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	return missingKeyZcardTestCase.Run(client, logger)
}
