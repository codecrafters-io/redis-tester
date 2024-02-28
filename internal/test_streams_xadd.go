package internal

import (
	"fmt"
	"math/rand"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/codecrafters-io/tester-utils/logger"
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

		return err
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

	strings := [10]string{
		"hello",
		"world",
		"mangos",
		"apples",
		"oranges",
		"watermelons",
		"grapes",
		"pears",
		"horses",
		"elephants",
	}

	randomKey := strings[rand.Intn(10)]

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
	}

	return nil
}
