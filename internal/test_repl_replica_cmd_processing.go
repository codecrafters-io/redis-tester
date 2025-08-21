package internal

import (
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplCmdProcessing(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	logger := stageHarness.Logger
	defer logger.ResetSecondaryPrefixes()

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

	logger.UpdateLastSecondaryPrefix("handshake")

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

	receiveReplicationHandshakeTestCase := test_cases.ReceiveReplicationHandshakeTestCase{}

	if err := receiveReplicationHandshakeTestCase.RunAll(master, logger); err != nil {
		return err
	}

	logger.UpdateLastSecondaryPrefix("propagation")
	master.UpdateBaseLogger(logger)

	replicaClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6380", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer replicaClient.Close()

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}

	for i := 1; i <= len(kvMap); i++ { // We need order of commands preserved
		key, value := kvMap[i][0], kvMap[i][1]
		// We are propagating commands to Replica as Master, don't expect any response back.
		if err := master.SendCommand("SET", key, value); err != nil {
			return err
		}
	}

	logger.UpdateLastSecondaryPrefix("test")
	replicaClient.UpdateBaseLogger(logger)
	// Add a small delay to wait for commands to be propagated
	// This was done in https://github.com/codecrafters-io/redis-tester/pull/212 to adjust fixtures
	// when fixtures are recorded from docker container
	time.Sleep(time.Millisecond)

	for i := 1; i <= len(kvMap); i++ {
		key, value := kvMap[i][0], kvMap[i][1]
		logger.Infof("Getting key %s", key)
		getCommandTestCase := test_cases.SendCommandTestCase{
			Command:   "GET",
			Args:      []string{key},
			Assertion: resp_assertions.NewStringAssertion(value),
			Retries:   5,
			ShouldRetryFunc: func(value resp_value.Value) bool {
				return resp_assertions.NewNilAssertion().Run(value) == nil
			},
		}

		if err := getCommandTestCase.Run(replicaClient, logger); err != nil {
			return err
		}
	}

	return nil
}
