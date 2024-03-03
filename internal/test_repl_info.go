package internal

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplInfo(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

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

	if infoMap[key] == "" {
		return fmt.Errorf("Expected: 'role' key in INFO replication.")
	}

	if role != "master" {
		return fmt.Errorf("Expected: 'role' to be 'master' in INFO replication, got %v", role)
	}

	client.Close()
	return nil
}
