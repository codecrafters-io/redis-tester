package internal

import (
	"fmt"
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
)

// Tests 'GET, SET'
func testGetSet(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient()

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

	logger.Debugf("Setting key %s to %s", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 0).Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	}

	logger.Debugf("Getting key %s", randomKey)
	resp, err = client.Get(randomKey).Result()
	if err != nil {
		return err
	}

	if resp != randomValue {
		return fmt.Errorf("Expected %#v, got %#v", randomValue, resp)
	}

	client.Close()
	return nil
}
