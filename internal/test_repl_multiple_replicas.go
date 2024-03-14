package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_utils "github.com/codecrafters-io/redis-tester/internal/resp"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplMultipleReplicas(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	// Run the user's code as a master
	masterBinary := redis_executable.NewRedisExecutable(stageHarness)
	if err := masterBinary.Run([]string{
		"--port", "6379",
	}); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// We use one client to send commands to the master
	client, err := instrumented_resp_connection.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	replicaCount := 3
	// We use multiple replicas to assert whether sent commands are replicated from the master (user's code)
	replicas, err := SpawnReplicas(replicaCount, stageHarness, logger, "localhost:6379")
	if err != nil {
		return err
	}
	for _, replica := range replicas {
		defer replica.Close()
	}

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}

	// We send SET commands to the master in order (user's code)
	for i := 1; i <= len(kvMap); i++ {
		key, value := kvMap[i][0], kvMap[i][1]
		setCommandTestCase := test_cases.CommandTestCase{
			Command:   "SET",
			Args:      []string{key, value},
			Assertion: resp_assertions.NewStringAssertion("OK"),
		}
		if err := setCommandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// We then assert that across all the replicas we receive the SET commands in order
	for j := 0; j < replicaCount; j++ {
		replica := replicas[j]
		logger.Infof("Testing Replica : %v", j+1)
		for i := 1; i <= len(kvMap); i++ {
			receiveCommandTestCase := &test_cases.ReceiveValueTestCase{
				Assertion:                 resp_assertions.NewCommandAssertion("SET", kvMap[i][0], kvMap[i][1]),
				ShouldSkipUnreadDataCheck: i < len(kvMap), // Except in the last case, we're expecting more SET commands to be present
			}

			if err := receiveCommandTestCase.Run(replica, logger); err != nil {
				// Redis sends a SELECT command, but we don't expect it from users.
				// If the first command is a SELECT command, we'll re-run the test case to test the next command instead
				if i == 1 && resp_utils.IsSelectCommand(receiveCommandTestCase.ActualValue) {
					if err := receiveCommandTestCase.Run(replica, logger); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}

	return nil
}
