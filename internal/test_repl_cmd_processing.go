package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplCmdProcessing(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
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
	logger.Infof("Test : 12")

	addr := "localhost:6379"
	masterClient := NewRedisClient(addr)
	addr = "localhost:6380"
	replicaClient := NewRedisClient(addr)

	masterClient.Close()
	replicaClient.Close()
	return fmt.Errorf("Test not implemented.")
}
