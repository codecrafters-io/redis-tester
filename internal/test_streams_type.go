package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testStreamsType(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	logger.Debugln("Setting key some_key to foo")
	resp, err := client.Set("some_key", "foo", 0).Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected \"OK\", got %#v", resp)
	}

	logger.Debugln("Sending type command")
	resp, err = client.Type("some_key").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "foo" {
		return fmt.Errorf("Expected \"foo\", got %#v", resp)
	}

	return nil
}
