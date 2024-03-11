package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsType(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run([]string{}); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	randomValue := testerutils_random.RandomWord()

	logger.Infof("$ redis-cli set %q %q", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 0).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		logger.Infof("Received response: \"%q\"", resp)
		return fmt.Errorf("Expected \"OK\", got %q", resp)
	} else {
		logger.Successf("Received response: \"%q\"", resp)
	}

	logger.Infof("$ redis-cli type %q", randomKey)
	resp, err = client.Type(randomKey).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "string" {
		return fmt.Errorf("Expected \"string\", got %q", resp)
	} else {
		logger.Successf("Type of %q is %q", randomKey, resp)
	}

	logger.Infof("$ redis-cli type %q", "missing_key"+"_"+randomValue)
	resp, err = client.Type("missing_key" + "_" + randomValue).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "none" {
		return fmt.Errorf("Expected \"none\", got %q", resp)
	} else {
		logger.Successf("Type of missing_key_%q is %q", randomValue, resp)
	}

	return nil
}
