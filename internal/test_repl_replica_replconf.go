package internal

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplReplicaSendsReplconf(stageHarness *testerutils.StageHarness) error {
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

	actualMessages, _ := readRespMessages(r, logger)
	expectedMessages := []string{"PING"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("PING received.")
	arg := []byte("+PONG\r\n")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))

	actualMessages, _ = readRespMessages(r, logger)
	expectedMessages = []string{"REPLCONF", "listening-port", "6380"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("REPLCONF listening-port 6380 received.")

	arg = []byte("+OK\r\n")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))

	actualMessages, _ = readRespMessages(r, logger)
	expectedMessages = []string{"REPLCONF", "*", "*", "*", "*"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf(strings.Join(actualMessages, " ") + " received.")

	return nil
}
