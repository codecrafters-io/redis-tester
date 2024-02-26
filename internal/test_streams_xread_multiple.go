package internal

import (
	"fmt"
	"math/rand"
	"reflect"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXreadMultiple(stageHarness *testerutils.StageHarness) error {
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

	otherStrings := [10]string{
		"bananas",
		"kiwis",
		"cherries",
		"strawberries",
		"blueberries",
		"raspberries",
		"pineapples",
		"coconuts",
		"peaches",
		"plums",
	}

	randomKey := strings[rand.Intn(10)]
	otherRandomKey := otherStrings[rand.Intn(10)]

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	})

	testXadd(client, logger, XADDTest{
		streamKey:        otherRandomKey,
		id:               "0-2",
		values:           map[string]interface{}{"bar": "baz"},
		expectedResponse: "0-2",
	})

	logger.Infof("$ redis-cli xread streams %s %s 0-0 0-1", randomKey, otherRandomKey)

	resp, err := client.XRead(&redis.XReadArgs{
		Streams: []string{randomKey, otherRandomKey, "0-0", "0-1"},
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
					ID:     "0-1",
					Values: map[string]interface{}{"foo": "bar"},
				},
			},
		},
		{
			Stream: otherRandomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"bar": "baz"},
				},
			},
		},
	}

	if !(reflect.DeepEqual(resp, expectedResp)) {
		return fmt.Errorf("Expected %#v, got %#v", expectedResp, resp)
	}

	return nil
}
