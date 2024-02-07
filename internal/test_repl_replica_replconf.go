package internal

import (
	"fmt"
	"net"

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
	w := resp3.NewWriter(conn)

	err = readAndAssertMessages(r, []string{"PING"}, logger)
	if err != nil {
		return err
	}
	message := "+PONG\r\n"
	sendAndLogMessage(w, message, logger)

	err = readAndAssertMessages(r, []string{"REPLCONF", "listening-port", "6380"}, logger)
	if err != nil {
		return err
	}
	message = "+OK\r\n"
	sendAndLogMessage(w, message, logger)

	err = readAndAssertMessages(r, []string{"REPLCONF", "*", "*", "*", "*"}, logger)
	if err != nil {
		return err
	}

	return nil
}
