package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlock(stageHarness *testerutils.StageHarness) error {
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

	go func() {
		time.Sleep(500 * time.Millisecond)

		testXadd(client, logger, XADDTest{
			streamKey:        randomKey,
			id:               "0-2",
			values:           map[string]interface{}{"bar": "baz"},
			expectedResponse: "0-2",
		})
	}()

	logger.Infof("$ redis-cli xread block 1000 streams %s 0-1", randomKey)

	resp, err := client.XRead(&redis.XReadArgs{
		Block:   1000 * time.Millisecond,
		Streams: []string{randomKey, "0-1"},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"bar": "baz"},
				},
			},
		},
	}

	if !reflect.DeepEqual(resp, expectedResp) {
		logger.Infof("Received response: %#v", resp)
		return fmt.Errorf("Expected %#v, got %#v", expectedResp, resp)
	} else {
		logger.Successf("Received response: %#v", resp)
	}

	resp, err = client.XRead(&redis.XReadArgs{
		Block:   1000 * time.Millisecond,
		Streams: []string{randomKey, "0-2"},
	}).Result()

	if err.Error() != "redis: nil" {
		logFriendlyError(logger, err)
		return err
	}

	if resp != nil {
		logger.Infof("this is a test")
		return fmt.Errorf("Expected %#v, got %#v", []redis.XStream{}, resp)
	} else {
		logger.Successf("Received response: %#v", resp)
	}

	return nil
}
