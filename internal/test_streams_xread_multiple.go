package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"strconv"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

func testStreamsXreadMultiple(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	var otherRandomKey string

	for {
		otherRandomKey = testerutils_random.RandomWord()
		if otherRandomKey != randomKey {
			break
		}
	}

	randomInt := testerutils_random.RandomInt(1, 100)
	otherRandomInt := testerutils_random.RandomInt(1, 100)

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

	xaddTest = &XADDTest{
		streamKey:        otherRandomKey,
		id:               "0-2",
		values:           map[string]interface{}{"humidity": otherRandomInt},
		expectedResponse: "0-2",
	}

	err = xaddTest.Run(client, logger)

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
		{
			Stream: otherRandomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"humidity": strconv.Itoa(otherRandomInt)},
				},
			},
		},
	}

	xreadTest := &XREADTest{
		streams:          []string{randomKey, otherRandomKey, "0-0", "0-1"},
		expectedResponse: expectedResp,
	}

	err = xreadTest.Run(client, logger)

	if err != nil {
		return err
	}

	return nil
}
