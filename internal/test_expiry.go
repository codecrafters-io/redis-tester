package internal

import (
	"fmt"
	"math/rand"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

// Tests Expiry
func testExpiry(stageHarness *testerutils.StageHarness) error {
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

	logger.Debugf("Running command: set %s %s px 100", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 100*time.Millisecond).Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	if resp != "OK" {
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	}

	logger.Debugf("Running command: get %s", randomKey)
	resp, err = client.Get(randomKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("Expected %#v, got nil", randomValue)
		}
		logFriendlyError(logger, err)
		return err
	}
	if resp != randomValue {
		return fmt.Errorf("Expected %#v, got %#v", randomValue, resp)
	}

	logger.Debugf("Sleeping for 101ms")
	time.Sleep(101 * time.Millisecond)

	logger.Debugf("Running command: get %s", randomKey)
	resp, err = client.Get(randomKey).Result()
	if err != redis.Nil {
		if err == nil {
			logger.Debugf("Hint: Read about null bulk strings in the Redis protocol docs")
			return fmt.Errorf("Expected null string, got %#v", resp)
		}

		logFriendlyError(logger, err)
		return err
	}

	client.Close()
	return nil
}
