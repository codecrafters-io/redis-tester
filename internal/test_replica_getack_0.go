package internal

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplGetaAckZero(stageHarness *testerutils.StageHarness) error {
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

	arg = []byte("*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*")
	conn.Write(arg)
	logger.Infof("%s sent.", bytes.TrimSpace(arg))
	actualMessages, err = readRespMessages(r, logger)
	expectedMessages = []string{"REPLCONF", "ACK", "0"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("REPLCONF ACK 0 received.")

	return nil
}
