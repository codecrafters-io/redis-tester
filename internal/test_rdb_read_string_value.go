package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbReadStringValue(stageHarness *test_case_harness.TestCaseHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	randomKeyAndValue := testerutils_random.RandomWords(2)
	randomKey := randomKeyAndValue[0]
	randomValue := randomKeyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: randomKey, value: randomValue}}); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with single key-value pair: %s=%q", randomKey, randomValue)

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run([]string{
		"--dir", RDBFileCreator.Dir,
		"--dbfilename", RDBFileCreator.Filename,
	}...); err != nil {
		return err
	}

	client := NewRedisClient("localhost:6379")

	logger.Infof(fmt.Sprintf("$ redis-cli GET %s", randomKey))
	resp, err := client.Get(randomKey).Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != randomValue {
		return fmt.Errorf("Expected response to be %v, got %v", randomValue, resp)
	}

	client.Close()
	return nil
}
