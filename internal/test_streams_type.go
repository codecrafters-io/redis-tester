package internal

import (
	"fmt"
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testStreamsType(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	strings := [10]string{
		"hello",
		"world",
		"mangos",
		"apples",
		"oranges",
		"watermelons",
		"grapes",
		"pears",
		"horses",
		"elephants",
	}

	randomKey := strings[rand.Intn(10)]
	randomValue := strings[rand.Intn(10)]

	logger.Infof("$ redis-cli set %s %s", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 0).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	}

	logger.Infof("$ redis-cli type %s", randomKey)
	resp, err = client.Type(randomKey).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "string" {
		return fmt.Errorf("Expected \"string\", got %#v", resp)
	}

	logger.Infof("$ redis-cli type %s", "missing_key"+"_"+randomValue)
	resp, err = client.Type("missing_key" + "_" + randomValue).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "none" {
		return fmt.Errorf("Expected \"none\", got %#v", resp)
	}

	return nil
}
