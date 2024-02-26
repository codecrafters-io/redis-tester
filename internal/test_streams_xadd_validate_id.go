package internal

import (
	"fmt"
	"math/rand"

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

func testXADD(client *redis.Client, logger *logger.Logger, test XADDTest) error {
	logger.Infof("$ redis-cli xadd %s %s %s", test.streamKey, test.id, test.values)

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: test.streamKey,
		ID:     test.id,
		Values: test.values,
	}).Result()

	logger.Infof("Received response: %s", resp)

	if err != nil {
		logFriendlyError(logger, err)
		return err
	} else {
		if test.expectedError == "" {
			return err
		}

		if err.Error() != test.expectedError {
			return fmt.Errorf("Expected %#v, got %#v", test.expectedError, err.Error())
		}
	}

	if resp != test.expectedResponse {
		return fmt.Errorf("Expected %#v, got %#v", test.expectedResponse, resp)
	}

	logger.Successf("Successfully added entry to stream")

	return nil
}

func testStreamsXaddValidateId(stageHarness *testerutils.StageHarness) error {
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

	tests := []XADDTest{
		{streamKey: randomKey, id: "1-1", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "1-1", expectedError: ""},
		{streamKey: randomKey, id: "1-2", values: map[string]interface{}{"bar": "baz"}, expectedResponse: "1-2", expectedError: ""},
		{streamKey: randomKey, id: "1-2", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-3", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-0", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD must be greater than 0-0"},
	}

	for _, test := range tests {
		testXADD(client, logger, test)
	}

	return nil
}
