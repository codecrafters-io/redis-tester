package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplReplicationID(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
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

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	var idKey, offsetKey string
	idKey, offsetKey = "master_replid", "master_repl_offset"

	if infoMap[idKey] == "" {
		return fmt.Errorf("Expected: '%v' key in INFO replication.", idKey)
	}

	if infoMap[offsetKey] != "0" {
		return fmt.Errorf("Expected: 0 value for '%v' in INFO replication.", offsetKey)
	}

	client.Close()
	return nil
}
