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

func testXadd(client *redis.Client, logger *logger.Logger, test XADDTest) error {
	var values []string

	for key, value := range test.values {
		values = append(values, key, fmt.Sprintf("%v", value))
	}

	logger.Infof("$ redis-cli xadd %s %s %s", test.streamKey, test.id, strings.Join(values, " "))

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: test.streamKey,
		ID:     test.id,
		Values: test.values,
	}).Result()

	if err != nil && test.expectedError == "" {
		logFriendlyError(logger, err)
		return err
	}

	if err != nil && test.expectedError != "" {
		if err.Error() != test.expectedError {
			return fmt.Errorf("Expected %#v, got %#v", test.expectedError, err.Error())
		}

		logger.Successf("Received error: \"%s\"", err.Error())
		return nil
	}

	if resp != test.expectedResponse {
		logger.Infof("Received response: \"%s\"", resp)
		return fmt.Errorf("Expected %#v, got %#v", test.expectedResponse, resp)
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

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	})

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
