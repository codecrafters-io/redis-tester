package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlock(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	randomInt := testerutils_random.RandomInt(1, 100)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-1",
	})

	var resp []redis.XStream
	var err error

	done := make(chan bool)

	go func() error {
		logger.Infof("$ redis-cli xread block %v streams %s", 1000, strings.Join([]string{randomKey, "0-1"}, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: []string{randomKey, "0-1"},
			Block:   1000,
		}).Result()

		if err != nil {
			logFriendlyError(logger, err)
			return err
		}

		done <- true
		return nil
	}()

	time.Sleep(500 * time.Millisecond)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-2",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-2",
	})

	<-done

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"temperature": strconv.Itoa(randomInt)},
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

	logger.Infof("$ redis-cli xread block %v streams %s", 1000, strings.Join([]string{randomKey, "0-2"}, " "))

	resp, err = client.XRead(&redis.XReadArgs{
		Streams: []string{randomKey, "0-2"},
		Block:   1000,
	}).Result()

	if err != redis.Nil {
		if err == nil {
			logger.Debugf("Hint: Read about null bulk strings in the Redis protocol docs")
			return fmt.Errorf("Expected null string, got %#v", resp)
		}

		logFriendlyError(logger, err)
		return err
	}

	return nil
}
