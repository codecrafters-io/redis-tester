package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

func testStreamsXrangeMinId(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
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

	maxId := "0-" + strconv.Itoa(randomNumber-1)

	for i := 1; i <= randomNumber-1; i++ {
		id := "0-" + strconv.Itoa(i)

		expectedResp = append(expectedResp, redis.XMessage{
			ID: id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		})
	}

	logger.Infof("$ redis-cli xrange %q - %q", randomKey, maxId)
	resp, err := client.XRange(randomKey, "-", maxId).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
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
		logger.Infof("Received response: \"%q\"", string(respJson))
		return fmt.Errorf("Expected %q, got %q", string(expectedRespJson), string(respJson))
	} else {
		logger.Successf("Received response: \"%q\"", string(respJson))
	}

	return nil
}
