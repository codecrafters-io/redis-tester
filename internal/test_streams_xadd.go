package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testStreamsXadd(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	logger.Debugln("Sending XADD command")
	logger.Infoln("$ redis-cli xadd stream_key 0-1 foo bar")

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: "stream_key",
		ID:     "0-1",
		Values: map[string]interface{}{
			"foo": "bar",
		},
	}).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "0-1" {
		return fmt.Errorf("Expected \"0-1\", got %#v", resp)
	}

	logger.Debugln("Sending type command with added stream key")
	logger.Infoln("$ redis-cli type stream_key")
	resp, err = client.Type("stream_key").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "stream" {
		return fmt.Errorf("Expected \"stream\", got %#v", resp)
	}

	return nil
}
