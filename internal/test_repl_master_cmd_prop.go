package internal

import (
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_client"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplMasterCmdProp(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()

	// Run the user's code as a master
	masterBinary := NewRedisBinary(stageHarness)
	masterBinary.args = []string{
		"--port", "6379",
	}

	if err := masterBinary.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// We use one client to send commands to the master
	client, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	// We use another client to assert whether sent commands are replicated from the master (user's code)
	replicaClient, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "replica")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}
	if err := sendHandshakeTestCase.RunAll(replicaClient, logger); err != nil {
		return err
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

	// We then assert that as a replica we receive the SET commands in order
	for i := 1; i <= len(kvMap); i++ {
		receiveCommandTestCase := &test_cases.ReceiveValueTestCase{
			Assertion:                 resp_assertions.NewCommandAssertion("SET", kvMap[i][0], kvMap[i][1]),
			ShouldSkipUnreadDataCheck: i < len(kvMap), // Except in the last case, we're expecting more SET commands to be present
		}

		if err := receiveCommandTestCase.Run(replicaClient, logger); err != nil {
			// Redis sends a SELECT command, but we don't expect it from users.
			// If the first command is a SELECT command, we'll re-run the test case to test the next command instead
			if i == 1 && isSelectCommand(receiveCommandTestCase.ActualValue) {
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

func isSelectCommand(value resp_value.Value) bool {
	return value.Type == resp_value.ARRAY &&
		len(value.Array()) > 0 &&
		value.Array()[0].Type == resp_value.BULK_STRING &&
		strings.ToLower(value.Array()[0].String()) == "select"
}
