package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

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

	logger.Infof("Received response: \"%s\"", resp)

	parts := strings.Split(resp, "-")

	if len(parts) != 2 {
		return fmt.Errorf("Expected a string in the form \"<millisecondsTime>-<sequenceNumber>\", got %s", resp)
	}

	timeStr, sequenceNumber := parts[0], parts[1]
	timeInt64, _ := strconv.ParseInt(timeStr, 10, 64)
	now := time.Now().Unix() * 1000
	oneSecondAgo := now - 1000
	oneSecondLater := now + 1000

	if len(timeStr) != 13 {
		return fmt.Errorf("Expected the first part of the ID to be a unix timestamp (%d characters), got %d characters", len(strconv.FormatInt(now, 10)), len(timeStr))
	} else if !(timeInt64 > oneSecondAgo && timeInt64 < oneSecondLater) {
		return fmt.Errorf("Expected the first part of the ID to be a valid unix timestamp, got %s", timeStr)
	} else {
		logger.Successf("The first part of the ID is a valid unix milliseconds timestamp")
	}

	if sequenceNumber != "0" {
		return fmt.Errorf("Expected the second part of the ID to be a sequence number with a value of \"0\", got %#v", sequenceNumber)
	} else {
		logger.Successf("The second part of the ID is a valid sequence number")
	}

	return nil
}
