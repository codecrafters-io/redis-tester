package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplMasterReplconf(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient()

	logger.Infof("$ redis-cli PING")
	resp, err := client.Do("PING").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "PONG" {
		return fmt.Errorf("Expected OK from Master, received %v", resp)
	}
	logger.Successf("PONG received.")

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")
	resp, err = client.Do("REPLCONF", "listening-port", "6380").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected OK from Master, received %v", resp)
	}

	client.Close()
	return nil
}
