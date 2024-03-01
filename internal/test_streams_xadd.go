package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
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

	logger.Infof("$ redis-cli xadd %s %s %s", t.streamKey, t.id, strings.Join(values, " "))

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
			return fmt.Errorf("Expected %#v, got %#v", t.expectedError, err.Error())
		}

		logger.Successf("Received error: \"%s\"", err.Error())
		return nil
	}

	if resp != t.expectedResponse {
		logger.Infof("Received response: \"%s\"", resp)
		return fmt.Errorf("Expected %#v, got %#v", t.expectedResponse, resp)
	} else {
		logger.Successf("Received response: \"%s\"", resp)
	}

	return nil
}

func testStreamsXadd(stageHarness *testerutils.StageHarness) error {
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

	logger.Infof("$ redis-cli type %s", randomKey)
	resp, err := client.Type(randomKey).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "stream" {
		return fmt.Errorf("Expected \"stream\", got %#v", resp)
	} else {
		logger.Successf("Type of %s is %s", randomKey, resp)
	}

	return nil
}
