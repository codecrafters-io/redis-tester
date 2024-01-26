package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func parseInfoOutput(lines []string, seperator string) map[string]string {
	infoMap := make(map[string]string)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		parts := strings.Split(trimmedLine, seperator)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			infoMap[key] = value
		}
	}
	return infoMap
}

func testReplReplicationID(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient()

	logger.Infof("$ redis-cli INFO replication")
	resp, err := client.Info("replication").Result()
	lines := strings.Split(resp, "\n")
	infoMap := parseInfoOutput(lines, ":")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	var idKey, offsetKey string
	idKey, offsetKey = "master_replid", "master_repl_offset"

	if infoMap[idKey] == "" {
		return fmt.Errorf("Expected '%v' key in INFO replication.", idKey)
	}

	if infoMap[offsetKey] != "0" {
		return fmt.Errorf("Expected 0 value for '%v' in INFO replication.", offsetKey)
	}

	client.Close()
	return nil
}
