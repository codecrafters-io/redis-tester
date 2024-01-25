package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplInfo(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient()

	logger.Infof("$ redis-cli INFO replication")
	resp, err := client.Info("replication").Result()
	roleLine := strings.TrimSpace(strings.Split(resp, "\n")[1])
	logger.Debugf("Role line is %s", roleLine)

	parts := strings.Split(roleLine, ":")
	var key, role string
	key, role = parts[0], parts[1]

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if key != "role" {
		return fmt.Errorf("Expected 'role' key in INFO replication, got %v", key)
	}

	if role != "master" {
		return fmt.Errorf("Expected 'role' to be 'master' in INFO replication, got %v", role)
	}

	client.Close()
	return nil
}
