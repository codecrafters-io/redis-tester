package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_utils "github.com/codecrafters-io/redis-tester/internal/resp"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
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
	defer listener.Close()

	logger.Infof("Master is running on port 6379")

	replica := redis_executable.NewRedisExecutable(stageHarness)
	if err := replica.Run([]string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}
	defer conn.Close()

	master, err := instrumented_resp_connection.NewInstrumentedRespConnection(stageHarness, conn, "master")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	receiveReplicationHandshakeTestCase := test_cases.ReceiveReplicationHandshakeTestCase{}

	if err := receiveReplicationHandshakeTestCase.RunAll(master, logger); err != nil {
		return err
	}

	offset := 0
	replicationTestCase := test_cases.ReplicationTestCase{}
	if err := replicationTestCase.RunGetAck(master, logger, offset); err != nil {
		return err
	}
	// How to handle offset in TestCase without breaking encapsulation ? ToDo Paul.
	offset += resp_utils.GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	command := append([]string{"PING"}, []string{}...)
	respValue := resp_value.NewStringArrayValue(command)
	if err := master.SendCommand(respValue); err != nil {
		return err
	}
	offset += resp_utils.GetByteOffset(command)

	if err := replicationTestCase.RunGetAck(master, logger, offset); err != nil {
		return err
	}
	offset += resp_utils.GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	key := resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value := resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	command = []string{"SET", key, value}
	respValue = resp_value.NewStringArrayValue(command)
	if err := master.SendCommand(respValue); err != nil {
		return err
	}
	offset += resp_utils.GetByteOffset(command)

	key = resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value = resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	command = []string{"SET", key, value}
	respValue = resp_value.NewStringArrayValue(command)
	if err := master.SendCommand(respValue); err != nil {
		return err
	}
	offset += resp_utils.GetByteOffset(command)

	return replicationTestCase.RunGetAck(master, logger, offset)
}
