package internal

import (
	"fmt"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplReplicaSendsPing(stageHarness *testerutils.StageHarness) error {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
	}
	defer listener.Close()
	logger := stageHarness.Logger

	logger.Infof("Server is running on port 6379.")

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
	return nil
}
