package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

func testStreamsXreadMultiple(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	otherRandomKey := testerutils_random.RandomWord()
	randomInt := testerutils_random.RandomInt(1, 100)
	otherRandomInt := testerutils_random.RandomInt(1, 100)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-1",
	})

	testXadd(client, logger, XADDTest{
		streamKey:        otherRandomKey,
		id:               "0-2",
		values:           map[string]interface{}{"humidity": otherRandomInt},
		expectedResponse: "0-2",
	})

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-1",
					Values: map[string]interface{}{"temperature": randomInt},
				},
			},
		},
		{
			Stream: otherRandomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"humidity": otherRandomInt},
				},
			},
		},
	}

	xreadTest := &XREADTest{
		streams:          []string{randomKey, otherRandomKey, "0-0", "0-1"},
		expectedResponse: expectedResp,
	}

	err := xreadTest.Run(client, logger)

	if err != nil {
		return err
	}

	return nil
}
