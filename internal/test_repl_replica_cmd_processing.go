package internal

import (
	"fmt"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplCmdProcessing(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
	}

	logger := stageHarness.Logger

	logger.Infof("Server is running on port 6379")

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

	r := resp3.NewReader(conn)
	w := resp3.NewWriter(conn)

	master := FakeRedisMaster{
		Reader: r,
		Writer: w,
		Logger: logger,
	}

	err = master.Handshake()
	if err != nil {
		return err
	}

	replicaAddr := "localhost:6380"
	replicaClient := NewRedisClient(replicaAddr)

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}
	for i := 1; i <= len(kvMap); i++ { // We need order of commands preserved
		key, value := kvMap[i][0], kvMap[i][1]
		logger.Debugf("Setting key %s to %s", key, value)
		command := "*3\r\n$3\r\nSET\r\n$3\r\n" + key + "\r\n$3\r\n" + value + "\r\n"
		sendMessage(w, command)
	}

	for i := 1; i <= len(kvMap); i++ {
		key, value := kvMap[i][0], kvMap[i][1]
		logger.Debugf("Getting key %s", key)
		resp, err := replicaClient.Get(key).Result()
		if err != nil {
			return err
		}
		if resp != value {
			return fmt.Errorf("Expected %#v, got %#v", value, resp)
		}
		logger.Successf("Received %v", resp)
	}

	replicaClient.Close()
	conn.Close()
	listener.Close()
	return nil
}
