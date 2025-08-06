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
	sortedSet := ds.GenerateZsetWithRandomMembers(ds.ZsetMemberGenerationOption{
		Count:          testerutils_random.RandomInt(4, 8),
		SameScoreCount: 2,
	})
	members := sortedSet.GetMembers()

	shuffledMembers := testerutils_random.ShuffleArray(members)
	for i, m := range shuffledMembers {
		// Add members
		zaddTestCase := test_cases.ZaddTestCase{
			Key:                  zsetKey,
			Member:               m,
			ExpectedAddedMembers: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}

		// Randomly test for cardinality
		shouldCheckCardinality := testerutils_random.RandomInt(0, 2)
		if shouldCheckCardinality == 1 {
			zcardTestCase := test_cases.SendCommandTestCase{
				Command:   "ZCARD",
				Args:      []string{zsetKey},
				Assertion: resp_assertions.NewIntegerAssertion(i + 1),
			}
			if err := zcardTestCase.Run(client, logger); err != nil {
				return err
			}
		}
	}

	// Update an existing member
	memberToUpdate := members[testerutils_random.RandomInt(0, sortedSet.Size())]
	newScore := ds.GetRandomZSetScore()
	zaddTestCase := test_cases.ZaddTestCase{
		Key:                  zsetKey,
		Member:               ds.NewSortedSetMember(memberToUpdate.GetName(), newScore),
		ExpectedAddedMembers: 0,
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

	// Check the cardinality of missing key */
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))
	missingKeyZcardTestCase := test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{missingKey},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	return missingKeyZcardTestCase.Run(client, logger)
}
