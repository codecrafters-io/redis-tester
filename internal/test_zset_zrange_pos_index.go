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
	zsetSize := testerutils_random.RandomInt(4, 8)
	members := GenerateRandomZSetMembers(ZsetMemberGenerationOption{
		Count:          zsetSize,
		SameScoreCount: 2,
	})

	zaddTestCase := test_cases.NewZsetTestCase(zsetkey)
	for _, m := range members {
		zaddTestCase.AddMember(m.Name, m.Score)
	}
	if err := zaddTestCase.RunZaddAll(client, logger); err != nil {
		return err
	}

	middleIndex := testerutils_random.RandomInt(1, zsetSize-1)

	// usual test cases
	if err := zaddTestCase.RunZrange(client, logger, 0, middleIndex); err != nil {
		return err
	}
	if err := zaddTestCase.RunZrange(client, logger, middleIndex, zsetSize-1); err != nil {
		return err
	}
	if err := zaddTestCase.RunZrange(client, logger, 0, zsetSize-1); err != nil {
		return err
	}

	// start index > end index
	if err := zaddTestCase.RunZrange(client, logger, 1, 0); err != nil {
		return err
	}

	// end index out of bounds
	if err := zaddTestCase.RunZrange(client, logger, 0, zsetSize+2); err != nil {
		return err
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
