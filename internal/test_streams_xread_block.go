package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlock(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	randomInt := testerutils_random.RandomInt(1, 100)

	xaddTest := &XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-1",
	}

	err := xaddTest.Run(client, logger)

	if err != nil {
		return err
	}

	var resp []redis.XStream

	done := make(chan bool)

	go func() error {
		logger.Infof("$ redis-cli xread block %d streams %v", 1000, strings.Join([]string{randomKey, "0-1"}, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: []string{randomKey, "0-1"},
			Block:   1000 * time.Millisecond,
		}).Result()

		if err != nil {
			logFriendlyError(logger, err)
			return err
		}

		done <- true
		return nil
	}()

	time.Sleep(500 * time.Millisecond)

	xaddTest = &XADDTest{
		streamKey:        randomKey,
		id:               "0-2",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-2",
	}

	err = xaddTest.Run(client, logger)

	if err != nil {
		return err
	}

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

	expectedRespJSON, err := json.MarshalIndent(expectedResp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	respJSON, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, expectedResp) {
		logger.Infof("Received response: \"%v\"", string(respJSON))
		return fmt.Errorf("Expected %v, got %v", string(expectedRespJSON), string(respJSON))
	} else {
		logger.Successf("Received response: \"%v\"", string(respJSON))
	}

	logger.Infof("$ redis-cli xread block %d streams %v", 1000, strings.Join([]string{randomKey, "0-2"}, " "))

	resp, err = client.XRead(&redis.XReadArgs{
		Streams: []string{randomKey, "0-2"},
		Block:   1000 * time.Millisecond,
	}).Result()

	if err != redis.Nil {
		if err == nil {
			logger.Debugf("Hint: Read about null bulk strings in the Redis protocol docs")
			return fmt.Errorf("Expected null string, got %q", resp)
		}

		logFriendlyError(logger, err)
		return err
	} else {
		logger.Successf("Received nil response")
	}

	return nil
}
