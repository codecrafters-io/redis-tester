package internal

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

type XADDTest struct {
	streamKey        string
	id               string
	values           map[string]interface{}
	expectedResponse string
	expectedError    string
}

func (t *XADDTest) Run(client *redis.Client, logger *logger.Logger) error {
	var values []string

	for key, value := range t.values {
		values = append(values, key, fmt.Sprintf("%v", value))
	}

	logger.Infof("$ redis-cli xadd %q %q %q", t.streamKey, t.id, strings.Join(values, " "))

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: t.streamKey,
		ID:     t.id,
		Values: t.values,
	}).Result()

	if err != nil && t.expectedError == "" {
		logFriendlyError(logger, err)
		return err
	}

	if err != nil && t.expectedError != "" {
		if err.Error() != t.expectedError {
			return fmt.Errorf("Expected %q, got %q", t.expectedError, err.Error())
		}

		logger.Successf("Received error: \"%q\"", err.Error())
		return nil
	}

	if resp != t.expectedResponse {
		logger.Infof("Received response: \"%q\"", resp)
		return fmt.Errorf("Expected %q, got %q", t.expectedResponse, resp)
	} else {
		logger.Successf("Received response: \"%q\"", resp)
	}

	return nil
}

func testStreamsXadd(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	xaddTest := &XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	}

	err := xaddTest.Run(client, logger)

	if err != nil {
		return err
	}

	logger.Infof("$ redis-cli type %q", randomKey)
	resp, err := client.Type(randomKey).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "stream" {
		return fmt.Errorf("Expected \"stream\", got %q", resp)
	} else {
		logger.Successf("Type of %q is %q", randomKey, resp)
	}

	return nil
}
