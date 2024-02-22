package internal

import (
	"fmt"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplReplicaSendsReplconf(stageHarness *testerutils.StageHarness) error {
	logger := stageHarness.Logger

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logger.Errorf("Error starting TCP server: %v", err)
		logFriendlyBindError(logger, err)
		return err
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

	err = master.AssertPing()
	if err != nil {
		return err
	}

	err = master.AssertReplConfPort()
	if err != nil {
		return err
	}

	err = master.AssertReplConfCapa()
	if err != nil {
		return err
	}

	conn.Close()
	listener.Close()
	return nil
}
