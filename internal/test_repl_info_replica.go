package internal

import (
	"fmt"
	"net"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	loggerutils "github.com/codecrafters-io/tester-utils/logger"
)

func testReplInfoReplica(stageHarness *testerutils.StageHarness) error {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return err
	}
	defer listener.Close()
	logger := stageHarness.Logger

	logger.Infof("Master is running on port 6379")

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}

	if err := replica.Run(); err != nil {
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
