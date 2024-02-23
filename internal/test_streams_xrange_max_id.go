package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXrangeMaxId(stageHarness *testerutils.StageHarness) error {
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

	max := 5
	min := 3
	randomNumber := rand.Intn(max-min+1) + min
	expected := []redis.XMessage{}

	for i := 1; i <= randomNumber; i++ {
		id := "0-" + strconv.Itoa(i)

		logger.Infof("$ redis-cli xadd %s %s foo bar", randomKey, id)

		resp, err := client.XAdd(&redis.XAddArgs{
			Stream: randomKey,
			ID:     id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		}).Result()

		if err != nil {
			logFriendlyError(logger, err)
			return err
		}

		if resp != id {
			return fmt.Errorf("Expected \"%s\", got %#v", id, resp)
		}
	}

	for i := 2; i <= randomNumber; i++ {
		id := "0-" + strconv.Itoa(i)

		expected = append(expected, redis.XMessage{
			ID: id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		})
	}

	logger.Infof("$ redis-cli xrange %s 0-2 +", randomKey)
	resp, err := client.XRange(randomKey, "0-2", "+").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, expected) {
		return fmt.Errorf("Expected %#v, got %#v", expected, resp)
	}

	return nil
}
