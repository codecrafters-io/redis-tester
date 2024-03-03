package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplCmdProcessing(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	logger := stageHarness.Logger

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logFriendlyBindError(logger, err)
		return fmt.Errorf("Error starting TCP server: %v", err)
	}

	logger.Infof("Master is running on port 6379")

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}

	if err := replica.Run(); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}

	master := NewFakeRedisMaster(conn, logger)

	err = master.Handshake()
	if err != nil {
		return err
	}

	conn1, err := NewRedisConn("", "localhost:6380")
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}
	replicaClient := NewFakeRedisMaster(conn1, logger)

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}

	for i := 1; i <= len(kvMap); i++ { // We need order of commands preserved
		key, value := kvMap[i][0], kvMap[i][1]
		err = master.Send([]string{"SET", key, value})
		if err != nil {
			return err
		}
		// Master is propagating commands to Replica, don't expect any response back.
	}

	for i := 1; i <= len(kvMap); i++ {
		key, value := kvMap[i][0], kvMap[i][1]
		logger.Infof("Getting key %s", key)
		err = replicaClient.SendAndAssertString([]string{"GET", key}, value, true)
		if err != nil {
			return err
		}
	}

	conn.Close()
	listener.Close()
	return nil
}
