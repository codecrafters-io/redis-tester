package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"net"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplGetaAckNonZero(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	logger := stageHarness.Logger

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logFriendlyBindError(logger, err)
		return fmt.Errorf("Error starting TCP server: %v", err)
	}

	logger.Infof("Master is running on port 6379")

	replica := redis_executable.NewRedisExecutable(stageHarness)
	if err := replica.Run([]string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}...); err != nil {
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

	offset := 0
	err = master.GetAck(offset) // 37
	if err != nil {
		return err
	}
	offset += GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	err = master.GetAck(offset) // 37
	if err != nil {
		return err
	}
	offset += GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	cmd := []string{"PING"}
	master.Send(cmd) // 14
	offset += GetByteOffset(cmd)

	err = master.GetAck(offset) // 37
	if err != nil {
		return err
	}
	offset += GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	key := RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value := RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	cmd = []string{"SET", key, value}
	master.Send(cmd) // 31
	offset += GetByteOffset(cmd)

	key = RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value = RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	cmd = []string{"SET", key, value}
	master.Send(cmd) // 31
	offset += GetByteOffset(cmd)

	err = master.GetAck(offset)
	if err != nil {
		return err
	}

	listener.Close()
	return nil
}
