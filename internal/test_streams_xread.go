package internal

import (
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXread(stageHarness *testerutils.StageHarness) error {
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

	logger.Infof("$ redis-cli xread streams %s 0-0", randomKey)
	resp := client.XRead(&redis.XReadArgs{
		Streams: []string{randomKey, "0-0"},
	})

	logger.Infof("resp: %v", resp)

	return nil
}
