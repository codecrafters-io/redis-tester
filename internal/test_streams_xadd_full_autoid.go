package internal

import (
	"fmt"
	"math/rand"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXaddFullAutoid(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	stringsList := [10]string{
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

	randomKey := stringsList[rand.Intn(10)]

	logger.Infof("$ redis-cli xadd %s * foo bar", randomKey)

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: randomKey,
		ID:     "*",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	parts := strings.Split(resp, "-")

	if len(parts) != 2 {
		return fmt.Errorf("Expected 2 parts, got %d", len(parts))
	}

	time, sequenceNumber := parts[0], parts[1]

	if len(time) != 13 {
		return fmt.Errorf("Expected 13 characters, got %d", len(time))
	}

	if sequenceNumber != "0" {
		return fmt.Errorf("Expected \"0\", got %#v", sequenceNumber)
	}

	return nil
}
