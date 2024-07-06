package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplGetaAckNonZero(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	logger := stageHarness.Logger
	defer logger.ResetSecondaryPrefix()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		logFriendlyBindError(logger, err)
		return fmt.Errorf("Error starting TCP server: %v", err)
	}
	defer listener.Close()

	logger.Infof("Master is running on port 6379")

	replica := redis_executable.NewRedisExecutable(stageHarness)
	if err := replica.Run("--port", "6380",
		"--replicaof", "localhost 6379"); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}
	defer conn.Close()

	master, err := instrumented_resp_connection.NewFromConn(logger, conn, "master")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.UpdateSecondaryPrefix("handshake")

	receiveReplicationHandshakeTestCase := test_cases.ReceiveReplicationHandshakeTestCase{}

	if err := receiveReplicationHandshakeTestCase.RunAll(master, logger); err != nil {
		return err
	}

	// The bytes received and sent during the handshake don't count towards offset.
	// After finishing the handshake we reset the counters.
	master.ResetByteCounters()
	logger.UpdateSecondaryPrefix("test")

	getAckTestCase := test_cases.GetAckTestCase{}
	if err := getAckTestCase.Run(master, logger, master.SentBytes); err != nil {
		return err
	}

	logger.UpdateSecondaryPrefix("propagation")
	if err := master.SendCommand("PING"); err != nil {
		return err
	}

	logger.UpdateSecondaryPrefix("test")
	if err := getAckTestCase.Run(master, logger, master.SentBytes); err != nil {
		return err
	}

	logger.UpdateSecondaryPrefix("propagation")
	key := testerutils_random.RandomWord()
	value := testerutils_random.RandomWord()
	if err := master.SendCommand("SET", []string{key, value}...); err != nil {
		return err
	}

	key = testerutils_random.RandomWord()
	value = testerutils_random.RandomWord()
	if err := master.SendCommand("SET", []string{key, value}...); err != nil {
		return err
	}

	logger.UpdateSecondaryPrefix("test")
	return getAckTestCase.Run(master, logger, master.SentBytes)
}
