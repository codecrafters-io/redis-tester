package internal

import (
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXreadMultiple(stageHarness *testerutils.StageHarness) error {
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

	otherStrings := [10]string{
		"bananas",
		"kiwis",
		"cherries",
		"strawberries",
		"blueberries",
		"raspberries",
		"pineapples",
		"coconuts",
		"peaches",
		"plums",
	}

	randomKey := strings[rand.Intn(10)]
	otherRandomKey := otherStrings[rand.Intn(10)]

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"foo": "bar"},
		expectedResponse: "0-1",
	})

	testXadd(client, logger, XADDTest{
		streamKey:        otherRandomKey,
		id:               "0-2",
		values:           map[string]interface{}{"bar": "baz"},
		expectedResponse: "0-2",
	})

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-1",
					Values: map[string]interface{}{"foo": "bar"},
				},
			},
		},
		{
			Stream: otherRandomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"bar": "baz"},
				},
			},
		},
	}

	testXread(client, logger, XREADTest{
		streams:          []string{randomKey, otherRandomKey, "0-0", "0-1"},
		expectedResponse: expectedResp,
	})

	return nil
}
