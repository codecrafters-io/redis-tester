package internal

import (
	"fmt"
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXaddValidateId(stageHarness *testerutils.StageHarness) error {
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

	logger.Infof("$ redis-cli xadd %s 1-1 foo bar", randomKey)

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "1-1",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "1-1" {
		return fmt.Errorf("Expected \"1-1\", got %#v", resp)
	}

	logger.Infof("$ redis-cli xadd %s 1-2 bar baz", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "1-2",
		Values: map[string]interface{}{
			"bar": "baz",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "1-2" {
		return fmt.Errorf("Expected \"1-2\", got %#v", resp)
	}

	logger.Infof("$ redis-cli xadd %s 1-2 baz foo", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "1-2",
		Values: map[string]interface{}{
			"baz": "foo",
		},
	}).Result()

	expectedErr := "ERR The ID specified in XADD is equal or smaller than the target stream top item"

	if err.Error() != expectedErr {
		return fmt.Errorf("Expected %#v, got %#v", expectedErr, err.Error())
	}

	logger.Infof("$ redis-cli xadd %s 0-3 baz foo", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "0-3",
		Values: map[string]interface{}{
			"baz": "foo",
		},
	}).Result()

	if err.Error() != expectedErr {
		return fmt.Errorf("Expected %#v, got %#v", expectedErr, err.Error())
	}

	logger.Infof("$ redis-cli xadd %s 0-0 baz foo", randomKey)

	resp, err = client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "0-0",
		Values: map[string]interface{}{
			"baz": "foo",
		},
	}).Result()

	expectedErr = "ERR The ID specified in XADD must be greater than 0-0"

	if err.Error() != expectedErr {
		return fmt.Errorf("Expected %#v, got %#v", expectedErr, err.Error())
	}

	return nil
}
