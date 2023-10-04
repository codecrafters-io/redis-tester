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

	logger.Infof("$ redis-cli set %s %s px 100", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 100*time.Millisecond).Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	if resp != "OK" {
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	}
	logger.Successf("Received OK (at %s)", time.Now().Format("15:04:05.000"))

	logger.Infof("$ redis-cli get %s (sent at %s, key should not be expired)", randomKey, time.Now().Format("15:04:05.000"))

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

	logger.Successf("Received %#v", randomValue)

	logger.Debugf("Sleeping for 101ms")
	time.Sleep(101 * time.Millisecond)

	logger.Infof("$ redis-cli get %s (sent at %s, key should be expired)", randomKey, time.Now().Format("15:04:05.000"))
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
