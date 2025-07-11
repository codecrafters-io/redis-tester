package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplMasterCmdProp(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	logger := stageHarness.Logger
	defer logger.ResetSecondaryPrefixes()

	// Run the user's code as a master
	masterBinary := redis_executable.NewRedisExecutable(stageHarness)
	if err := masterBinary.Run("--port", "6379"); err != nil {
		return err
	}

	logger.UpdateLastSecondaryPrefix("handshake")

	// We use one client to send commands to the master
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	// We use another client to assert whether sent commands are replicated from the master (user's code)
	replicaClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "replica")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer replicaClient.Close()

	client.UpdateBaseLogger(logger)
	replicaClient.UpdateBaseLogger(logger)

	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}
	if err := sendHandshakeTestCase.RunAll(replicaClient, logger, 6380); err != nil {
		return err
	}

	logger.UpdateLastSecondaryPrefix("test")
	client.UpdateBaseLogger(logger)
	replicaClient.UpdateBaseLogger(logger)

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}

	// We send SET commands to the master in order (user's code)
	for i := 1; i <= len(kvMap); i++ {
		key, value := kvMap[i][0], kvMap[i][1]

		setCommandTestCase := test_cases.SendCommandTestCase{
			Command:   "SET",
			Args:      []string{key, value},
			Assertion: resp_assertions.NewStringAssertion("OK"),
		}

		if err := setCommandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	logger.Successf("Sent 3 SET commands to master successfully.")

	// We then assert that as a replica we receive the SET commands in order
	for i := 1; i <= len(kvMap); i++ {
		replicaClient.GetLogger().Infof("Expecting \"SET %s %s\" to be propagated", kvMap[i][0], kvMap[i][1])

		receiveCommandTestCase := &test_cases.ReceiveValueTestCase{
			Assertion:                 resp_assertions.NewCommandAssertion("SET", kvMap[i][0], kvMap[i][1]),
			ShouldSkipUnreadDataCheck: i < len(kvMap), // Except in the last case, we're expecting more SET commands to be present
		}

		if err := receiveCommandTestCase.Run(replicaClient, logger); err != nil {
			// Redis sends a SELECT command, but we don't expect it from users.
			// If the first command is a SELECT command, we'll re-run the test case to test the next command instead
			if i == 1 && IsSelectCommand(receiveCommandTestCase.ActualValue) {
				if err := receiveCommandTestCase.Run(replicaClient, logger); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
