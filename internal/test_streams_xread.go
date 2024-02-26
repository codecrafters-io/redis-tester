package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/go-redis/redis"
)

type XREADTest struct {
	streams          []string
	block            *int
	expectedResponse []redis.XStream
	expectedError    error
}

func testXread(client *redis.Client, logger *logger.Logger, test XREADTest) error {
	logger.Infof("$ redis-cli xread streams %s", strings.Join(test.streams, " "))

	resp, err := client.XRead(&redis.XReadArgs{
		Streams: test.streams,
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, test.expectedResponse) {
		logger.Infof("Received response: \"%v\"", resp)
		return fmt.Errorf("Expected %#v, got %#v", test.expectedResponse, resp)
	} else {
		logger.Successf("Received response: \"%v\"", resp)
	}

	return nil
}

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
	}

	testXread(client, logger, XREADTest{
		streams:          []string{randomKey, "0-0"},
		expectedResponse: expectedResp,
	})

	return nil
}
