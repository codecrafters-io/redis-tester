package internal

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplCmdProcessing(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
	}
	defer listener.Close()
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

	actualMessages, err := readRespMessages(r, logger)
	if err != nil {
		return err
	}
	expectedMessages := []string{"PING"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("PING received.")
	arg := []byte("+PONG\r\n")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))

	actualMessages, err = readRespMessages(r, logger)
	if err != nil {
		return err
	}
	expectedMessages = []string{"REPLCONF", "listening-port", "6380"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("REPLCONF listening-port 6380 received.")
	arg = []byte("+OK\r\n")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))

	actualMessages, err = readRespMessages(r, logger)
	if err != nil {
		return err
	}
	expectedMessages = []string{"REPLCONF", "*", "*", "*", "*"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf(strings.Join(actualMessages, " ") + " received.")

	arg = []byte("+OK\r\n")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))

	actualMessages, err = readRespMessages(r, logger)
	if err != nil {
		return err
	}
	expectedMessages = []string{"PSYNC", "?", "-1"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("PSYNC ? -1 received.")

	arg = []byte("+FULLRESYNC c00d0def8c1d916ed06e2d2e69b8b658532a07ef 0\r\n")
	conn.Write(arg)
	response := SendRDBFile()
	w.Write(response)
	w.Flush()

	replicaAddr := "localhost:6380"
	replicaClient := NewRedisClient(replicaAddr)

	key1, value1 := "foo", "123"
	key2, value2 := "bar", "456"
	key3, value3 := "baz", "789"
	kvMap := map[int][]string{
		1: {key1, value1},
		2: {key2, value2},
		3: {key3, value3},
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

	return nil
}
