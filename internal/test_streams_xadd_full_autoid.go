package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

func testStreamsXaddFullAutoid(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run([]string{}); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	logger.Infof("$ redis-cli xadd %q * foo bar", randomKey)

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

	logger.Infof("Received response: \"%q\"", resp)

	parts := strings.Split(resp, "-")

	if len(parts) != 2 {
		return fmt.Errorf("Expected a string in the form \"<millisecondsTime>-<sequenceNumber>\", got %q", resp)
	}

	timeStr, sequenceNumber := parts[0], parts[1]
	timeInt64, _ := strconv.ParseInt(timeStr, 10, 64)
	now := time.Now().Unix() * 1000
	oneSecondAgo := now - 1000
	oneSecondLater := now + 1000

	if len(timeStr) != 13 {
		return fmt.Errorf("Expected the first part of the ID to be a unix timestamp (%d characters), got %d characters", len(strconv.FormatInt(now, 10)), len(timeStr))
	} else if !(timeInt64 > oneSecondAgo && timeInt64 < oneSecondLater) {
		return fmt.Errorf("Expected the first part of the ID to be a valid unix timestamp, got %q", timeStr)
	} else {
		logger.Successf("The first part of the ID is a valid unix milliseconds timestamp")
	}

	if sequenceNumber != "0" {
		return fmt.Errorf("Expected the second part of the ID to be a sequence number with a value of \"0\", got %q", sequenceNumber)
	} else {
		logger.Successf("The second part of the ID is a valid sequence number")
	}

	return nil
}
