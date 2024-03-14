package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"net"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplReplicaSendsPing(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logFriendlyBindError(logger, err)
		return fmt.Errorf("Error starting TCP server: %v", err)
	}

	logger.Infof("Master is running on port 6379.")

	replica := redis_executable.NewRedisExecutable(stageHarness)
	if err := replica.Run([]string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}...); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}

	master := NewFakeRedisMaster(conn, logger)

	err = master.AssertPing()
	if err != nil {
		return err
	}

	conn.Close()
	listener.Close()
	return nil
}
