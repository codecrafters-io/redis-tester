package internal

import (
	"fmt"
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXaddPartialAutoid(stageHarness *testerutils.StageHarness) error {
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

	logger.Infof("$ redis-cli xadd %s 0-* foo bar", randomKey)

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "0-*",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "0-1" {
		return fmt.Errorf("Expected \"0-1\", got %#v", resp)
	}

	logger.Infof("$ redis-cli xadd %s 1-* foo bar", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "1-*",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "1-0" {
		return fmt.Errorf("Expected \"1-0\", got %#v", resp)
	}

	logger.Infof("$ redis-cli xadd %s 1-* bar baz", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "1-*",
		Values: map[string]interface{}{
			"bar": "baz",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "1-1" {
		return fmt.Errorf("Expected \"1-1\", got %#v", resp)
	}

	return nil
}
