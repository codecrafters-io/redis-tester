package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZrangeNegIndex(stageHarness *test_case_harness.TestCaseHarness) error {
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

	startIndex := -zsetSize
	endIndex := -1
	middleIndex := zsetSize + testerutils_random.RandomInt(startIndex+1, -1)

	// usual test cases
	if err := zaddTestCase.RunZrange(client, logger, 0, middleIndex); err != nil {
		return err
	}
	if err := zaddTestCase.RunZrange(client, logger, middleIndex, endIndex); err != nil {
		return err
	}
	if err := zaddTestCase.RunZrange(client, logger, startIndex, endIndex); err != nil {
		return err
	}

	// start index > end index
	if err := zaddTestCase.RunZrange(client, logger, -1, -2); err != nil {
		return err
	}

	// end index out of bounds
	return zaddTestCase.RunZrange(client, logger, startIndex-1, endIndex)
}
