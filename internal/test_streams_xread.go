package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

type XREADTest struct {
	streams          []string
	block            *time.Duration
	expectedResponse []redis.XStream
}

func (t *XREADTest) Run(client *redis.Client, logger *logger.Logger) error {
	var resp []redis.XStream
	var err error

	if t.block == nil {
		logger.Infof("$ redis-cli xread streams %v", strings.Join(t.streams, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: t.streams,
			Block:   -1 * time.Millisecond, // Zero value for Block in XReadArgs struct is 0. Need a negative value to indicate no block.
		}).Result()
	} else {
		logger.Infof("$ redis-cli xread block %d streams %v", t.block.Milliseconds(), strings.Join(t.streams, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: t.streams,
			Block:   *t.block,
		}).Result()
	}

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	expectedRespJSON, err := json.MarshalIndent(t.expectedResponse, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	respJSON, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, t.expectedResponse) {
		logger.Infof("Received response: \"%v\"", string(respJSON))
		return fmt.Errorf("Expected %v, got %v", string(expectedRespJSON), string(respJSON))
	} else {
		logger.Successf("Received response: \"%v\"", string(respJSON))
	}

	return nil
}

func testStreamsXread(stageHarness *test_case_harness.TestCaseHarness) error {
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

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-1",
					Values: map[string]interface{}{"temperature": strconv.Itoa(randomInt)},
				},
			},
		},
	}

	xreadTest := &XREADTest{
		streams:          []string{randomKey, "0-0"},
		expectedResponse: expectedResp,
	}

	err = xreadTest.Run(client, logger)

	if err != nil {
		return err
	}

	return nil
}
