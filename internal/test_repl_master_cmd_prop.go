package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplMasterCmdProp(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	conn1, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewFakeRedisMaster(conn1, logger)

	replica := NewFakeRedisReplica(conn, logger)

	err = replica.Handshake()
	if err != nil {
		return err
	}

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}
	for i := 1; i <= len(kvMap); i++ { // We need order of commands preserved
		key, value := kvMap[i][0], kvMap[i][1]
		logger.Infof("Setting key %s to %s", key, value)
		client.Send([]string{"SET", key, value})
	}

	// Redis will send SELECT, but not expected from Users.
	err, _ = replica.readAndAssertMessagesWithSkip([]string{"SET", "foo", "123"}, "SELECT", true)

	if err != nil {
		return err
	}

	err, _ = replica.readAndAssertMessages([]string{"SET", "bar", "456"}, true)
	if err != nil {
		return err
	}

	err, _ = replica.readAndAssertMessages([]string{"SET", "baz", "789"}, true)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}
