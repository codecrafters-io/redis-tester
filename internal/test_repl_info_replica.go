package internal

import (
	"fmt"
	"net"
	"regexp"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"

	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	loggerutils "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplInfoReplica(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger

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

	go func(l net.Listener) error {
		// Connecting to master in this stage is optional.
		conn, err := listener.Accept()
		if err != nil {
			logger.Debugf("Error accepting: %s", err.Error())
			return err
		}
		defer conn.Close()

		quietLogger := loggerutils.GetQuietLogger("")
		master, err := instrumented_resp_connection.NewFromConn(quietLogger, conn, "master")
		if err != nil {
			logFriendlyError(quietLogger, err)
			return err
		}
		receiveReplicationHandshakeTestCase := test_cases.ReceiveReplicationHandshakeTestCase{}

		_ = receiveReplicationHandshakeTestCase.RunAll(master, quietLogger)

		return nil
	}(listener)

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6380", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "INFO",
		Args:                      []string{"replication"},
		Assertion:                 resp_assertions.NewNoopAssertion(),
		ShouldSkipUnreadDataCheck: true,
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	responseValue := commandTestCase.ReceivedResponse

	if responseValue.Type != resp_value.BULK_STRING && responseValue.Type != resp_value.SIMPLE_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", responseValue.Type)
	}

	var patternMatchError error

	if !regexp.MustCompile("role:").Match([]byte(responseValue.String())) {
		patternMatchError = fmt.Errorf("Expected role to be present in response. Got: %q", responseValue.String())
	}

	if regexp.MustCompile("role:slave").Match([]byte(responseValue.String())) {
		logger.Successf("Found role:slave in response.")
	} else {
		patternMatchError = fmt.Errorf("Expected role to be slave in response. Got: %q", responseValue.String())
	}

	return patternMatchError
}
