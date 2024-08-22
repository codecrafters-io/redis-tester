package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

func testStreamsXrangeMinID(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	max := 5
	min := 3
	randomNumber := testerutils_random.RandomInt(min, max)
	expectedResp := []redis.XMessage{}

	for i := 1; i <= randomNumber; i++ {
		id := "0-" + strconv.Itoa(i)

		xaddTest := &XADDTest{
			streamKey:        randomKey,
			id:               id,
			values:           map[string]interface{}{"foo": "bar"},
			expectedResponse: id,
		}

		err := xaddTest.Run(client, logger)

		if err != nil {
			return err
		}
	}

	maxID := "0-" + strconv.Itoa(randomNumber-1)

	for i := 1; i <= randomNumber-1; i++ {
		id := "0-" + strconv.Itoa(i)

		expectedResp = append(expectedResp, redis.XMessage{
			ID: id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		})
	}

	logger.Infof("$ redis-cli xrange %v - %v", randomKey, maxID)
	resp, err := client.XRange(randomKey, "-", maxID).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
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
