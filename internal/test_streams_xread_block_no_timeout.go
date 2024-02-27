package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlockNoTimeout(stageHarness *testerutils.StageHarness) error {
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

	respChan := make(chan *redis.XStream, 1)

	go func() error {
		resp, err := client.XRead(&redis.XReadArgs{
			Streams: []string{randomKey, "0-1"},
			Block:   0,
		}).Result()

		if err != nil {
			logger.Errorf("Error: %v", err)
			return err
		}

		respChan <- &resp[0]

		return nil
	}()

	time.Sleep(time.Second)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-2",
		values:           map[string]interface{}{"bar": "baz"},
		expectedResponse: "0-2",
	})

	resp := <-respChan

	expectedResp := &redis.XStream{
		Stream: randomKey,
		Messages: []redis.XMessage{
			{
				ID:     "0-2",
				Values: map[string]interface{}{"bar": "baz"},
			},
		},
	}

	if !reflect.DeepEqual(resp, expectedResp) {
		logger.Infof("Received response: \"%v\"", resp)
		return fmt.Errorf("Expected %#v, got %#v", expectedResp, resp)
	} else {
		logger.Successf("Received response: \"%v\"", resp)
	}

	return nil
}
