package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

func testStreamsXrange(stageHarness *testerutils.StageHarness) error {
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

		testXadd(client, logger, XADDTest{
			streamKey:        randomKey,
			id:               id,
			values:           map[string]interface{}{"foo": "bar"},
			expectedResponse: id,
		})

		expectedResp = append(expectedResp, redis.XMessage{
			ID: id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		})
	}

	maxId := "0-" + strconv.Itoa(randomNumber)
	expectedResp = expectedResp[1:]

	logger.Infof("$ redis-cli xrange %s 0-2 %s", randomKey, maxId)
	resp, err := client.XRange(randomKey, "0-2", maxId).Result()

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
		logger.Infof("Received response: \"%s\"", string(respJson))
		return fmt.Errorf("Expected %#v, got %#v", string(expectedRespJson), string(respJson))
	} else {
		logger.Successf("Received response: \"%s\"", string(respJson))
	}

	return nil
}
