package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlockNoTimeout(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	})

	respChan := make(chan *[]redis.XStream, 1)

	go func() error {
		logger.Infof("$ redis-cli block %v xread streams %s", 0, strings.Join([]string{randomKey, "0-1"}, " "))

		resp, err := client.XRead(&redis.XReadArgs{
			Streams: []string{randomKey, "0-1"},
			Block:   0,
		}).Result()

		if err != nil {
			logger.Errorf("Error: %v", err)
			return err
		}

		respChan <- &resp

		return nil
	}()

	time.Sleep(1000 * time.Millisecond)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-2",
		values:           map[string]interface{}{"bar": "baz"},
		expectedResponse: "0-2",
	})

	resp := <-respChan

	expectedResp := &[]redis.XStream{
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

	expectedRespJson, err := json.MarshalIndent(expectedResp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	respJson, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, expectedResp) {
		logger.Infof("Received response: \"%v\"", string(respJson))
		return fmt.Errorf("Expected %#v, got %#v", string(expectedRespJson), string(respJson))
	} else {
		logger.Successf("Received response: \"%v\"", string(respJson))
	}

	return nil
}
