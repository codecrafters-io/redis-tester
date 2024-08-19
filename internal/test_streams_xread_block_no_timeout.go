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

func testStreamsXreadBlockNoTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
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

	respChan := make(chan *[]redis.XStream, 1)

	go func() error {
		logger.Infof("$ redis-cli xread block %d streams %v", 0, strings.Join([]string{randomKey, "0-1"}, " "))

		resp, err := client.XRead(&redis.XReadArgs{
			Streams: []string{randomKey, "0-1"},
			Block:   0,
		}).Result()

		if err != nil {
			logger.Errorf("Error: %q", err)
			return err
		}

		respChan <- &resp

		return nil
	}()

	time.Sleep(1000 * time.Millisecond)

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

	resp := <-respChan

	expectedResp := &[]redis.XStream{
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

	return nil
}
