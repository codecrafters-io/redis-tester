package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplMasterReplconf(stageHarness *test_case_harness.TestCaseHarness) error {
	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run([]string{
		"--port", "6379",
	}...); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}

	if err := sendHandshakeTestCase.RunPingStep(client, logger); err != nil {
		return err
	}

	return sendHandshakeTestCase.RunReplconfStep(client, logger)
}
