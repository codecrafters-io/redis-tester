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

	zsetkey := testerutils_random.RandomWord()
	sortedSet := GenerateRandomSortedSet(SortedSetGenerationOption{Count: testerutils_random.RandomInt(4, 8), SameScoreCount: 2})

	for _, m := range sortedSet.Members() {
		zaddTestCase := test_cases.ZaddTestCase{
			Key:    zsetkey,
			Member: m,
		}

		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	middleIndex := testerutils_random.RandomInt(1, sortedSet.Size()-1)

	zrangeTestCases := []test_cases.ZrangeTestCase{
		// Happy-path cases
		{
			Key:                zsetkey,
			Start:              0,
			End:                middleIndex,
			ExpectedMemberKeys: sortedSet.MemberKeys()[0:middleIndex],
		},
		{
			Key:                zsetkey,
			Start:              middleIndex,
			End:                sortedSet.Size() - 1,
			ExpectedMemberKeys: sortedSet.MemberKeys()[middleIndex:],
		},
		// Start index > end index
		{
			Key:                zsetkey,
			Start:              1,
			End:                0,
			ExpectedMemberKeys: []string{},
		},
		// End index out of bounds
		{
			Key:                zsetkey,
			Start:              0,
			End:                sortedSet.Size() + 2,
			ExpectedMemberKeys: []string{},
		},
	}

	for _, zrangeTestCase := range zrangeTestCases {
		if err := zrangeTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// key does not exist
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))
	missingKeyZrangeTestCase := test_cases.SendCommandTestCase{
		Command:   "ZRANGE",
		Args:      []string{missingKey, "0", "1"},
		Assertion: resp_assertions.NewOrderedArrayAssertion(nil),
	}

	return missingKeyZrangeTestCase.Run(client, logger)
}
