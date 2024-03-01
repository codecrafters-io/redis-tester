package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testStreamsType(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	randomValue := testerutils_random.RandomWord()

	logger.Infof("$ redis-cli set %s %s", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 0).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		logger.Infof("Received response: \"%s\"", resp)
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	} else {
		logger.Successf("Received response: \"%s\"", resp)
	}

	logger.Infof("$ redis-cli type %s", randomKey)
	resp, err = client.Type(randomKey).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "string" {
		return fmt.Errorf("Expected \"string\", got %#v", resp)
	} else {
		logger.Successf("Type of %s is %s", randomKey, resp)
	}

	logger.Infof("$ redis-cli type %s", "missing_key"+"_"+randomValue)
	resp, err = client.Type("missing_key" + "_" + randomValue).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "none" {
		return fmt.Errorf("Expected \"none\", got %#v", resp)
	} else {
		logger.Successf("Type of missing_key_%s is %s", randomValue, resp)
	}

	return nil
}
