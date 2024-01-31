package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplMasterPsync(stageHarness *testerutils.StageHarness) error {
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
	logger.Successf("OK received.")

	logger.Infof("$ redis-cli PSYNC ? -1")
	resp, err = client.Do("PSYNC", "?", "-1").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	respStr, _ := resp.(string)
	respParts := strings.Split(respStr, " ")
	command := respParts[0]
	offset := respParts[2]

	if command != "FULLRESYNC" {
		return fmt.Errorf("Expected FULLRESYNC from Master, received %v", command)
	}
	logger.Successf("FULLRESYNC received.")

	if offset != "0" {
		return fmt.Errorf("Expected offset to be 0 from Master, received %v", offset)
	}
	logger.Successf("offset = 0 received.")

	client.Close()
	return nil
}
