package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_utils "github.com/codecrafters-io/redis-tester/internal/resp"
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
	}...); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}
	defer conn.Close()

	master, err := instrumented_resp_connection.NewFromConn(stageHarness, conn, "master")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	receiveReplicationHandshakeTestCase := test_cases.ReceiveReplicationHandshakeTestCase{}

	if err := receiveReplicationHandshakeTestCase.RunAll(master, logger); err != nil {
		return err
	}

	// Reset received and sent byte offset here.
	master.ResetByteCounters()

	replicationTestCase := test_cases.ReplicationTestCase{}
	if err := replicationTestCase.RunGetAck(master, logger, master.SentBytes); err != nil {
		return err
	}

	if err := master.SendCommand("PING"); err != nil {
		return err
	}

	if err := replicationTestCase.RunGetAck(master, logger, master.SentBytes); err != nil {
		return err
	}

	key := resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value := resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	if err := master.SendCommand("SET", []string{key, value}...); err != nil {
		return err
	}

	key = resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	value = resp_utils.RandomAlphanumericString(testerutils_random.RandomInt(5, 20))
	if err := master.SendCommand("SET", []string{key, value}...); err != nil {
		return err
	}

	return replicationTestCase.RunGetAck(master, logger, master.SentBytes)
}
