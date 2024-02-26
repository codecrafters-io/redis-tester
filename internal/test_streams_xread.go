package internal

import (
	"fmt"
	"math/rand"
	"reflect"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXread(stageHarness *testerutils.StageHarness) error {
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

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	})

	logger.Infof("$ redis-cli xread streams %s 0-0", randomKey)

	resp, err := client.XRead(&redis.XReadArgs{
		Streams: []string{randomKey, "0-0"},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Infof("Received response: \"%v\"", resp)

	expectedResp := map[string]interface{}{
		randomKey: []redis.XMessage{
			{
				ID:     "0-1",
				Values: map[string]interface{}{"foo": "bar"},
			},
		},
	}

	if !reflect.DeepEqual(resp, expectedResp) {
		return fmt.Errorf("Expected %#v, got %#v", expectedResp, resp)
	}

	return nil
}
