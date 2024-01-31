package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplInfoReplica(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6380",
	}

	if err := master.Run(); err != nil {
		return err
	}

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6379",
		"--replicaof", "localhost", "6380",
	}

	if err := replica.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient()

	logger.Infof("$ redis-cli INFO replication")
	resp, err := client.Info("replication").Result()
	lines := strings.Split(resp, "\n")
	infoMap := parseInfoOutput(lines, ":")
	key := "role"
	role := infoMap[key]

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if key != "role" {
		return fmt.Errorf("Expected 'role' key in INFO replication, got %v", key)
	}

	if role != "slave" {
		return fmt.Errorf("Expected 'role' to be 'slave' in INFO replication, got %v", role)
	}

	client.Close()
	return nil
}
