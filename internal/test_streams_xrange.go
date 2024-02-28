package internal

import (
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
	expected := []redis.XMessage{}

	for i := 1; i <= randomNumber; i++ {
		id := "0-" + strconv.Itoa(i)

		testXadd(client, logger, XADDTest{
			streamKey:        randomKey,
			id:               id,
			values:           map[string]interface{}{"foo": "bar"},
			expectedResponse: id,
		})

		expected = append(expected, redis.XMessage{
			ID: id,
			Values: map[string]interface{}{
				"foo": "bar",
			},
		})
	}

	maxId := "0-" + strconv.Itoa(randomNumber)
	expected = expected[1:]

	resp, err := client.XRange(randomKey, "0-2", maxId).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, expected) {
		logger.Infof("Received response: \"%s\"", resp)
		return fmt.Errorf("Expected %#v, got %#v", expected, resp)
	} else {
		logger.Successf("Received response: \"%s\"", resp)
	}

	return nil
}
