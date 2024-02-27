package internal

import (
	"math/rand"
	"sync"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXreadBlockMaxId(stageHarness *testerutils.StageHarness) error {
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

	var wg sync.WaitGroup
	wg.Add(1)

	blockDuration := 0 * time.Millisecond

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-2",
					Values: map[string]interface{}{"bar": "baz"},
				},
			},
		},
	}

	go func() {
		defer wg.Done()

		testXread(client, logger, XREADTest{
			streams:          []string{randomKey, "$"},
			block:            &blockDuration,
			expectedResponse: expectedResp,
		})
	}()

	time.Sleep(1000 * time.Millisecond)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-2",
		values:           map[string]interface{}{"bar": "baz"},
		expectedResponse: "0-2",
	})

	wg.Wait()

	blockDuration = 1000 * time.Millisecond

	testXread(client, logger, XREADTest{
		streams:          []string{randomKey, "$"},
		block:            &blockDuration,
		expectedResponse: []redis.XStream(nil),
		expectedError:    "redis: nil",
	})

	return nil
}
