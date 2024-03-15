package internal

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	loggerutils "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplInfoReplica(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logFriendlyBindError(logger, err)
		return fmt.Errorf("Error starting TCP server: %v", err)
	}

	defer listener.Close()
	logger.Infof("Master is running on port 6379")

	replica := redis_executable.NewRedisExecutable(stageHarness)
	if err := replica.Run("--port", "6380",
		"--replicaof", "localhost", "6379"); err != nil {
		return err
	}

	go func(l net.Listener) error {
		// Connecting to master in this stage is optional.
		conn, err := listener.Accept()
		if err != nil {
			logger.Debugf("Error accepting: %s", err.Error())
			return err
		}

		quietLogger := loggerutils.GetQuietLogger("")
		master := NewFakeRedisMaster(conn, quietLogger)
		master.Handshake()

		conn.Close()
		return nil
	}(listener)

	client := NewRedisClient("localhost:6380")

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
		return fmt.Errorf("Expected: 'role' and actual: '%v' keys in INFO replication don't match", key)
	}

	if role != "slave" {
		return fmt.Errorf("Expected: 'role' and actual: '%v' roles in INFO replication don't match", role)
	}

	client.Close()
	return nil
}
