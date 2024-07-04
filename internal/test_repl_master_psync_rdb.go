package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplMasterPsyncRdb(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()
	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run("--port", "6379"); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}

	return sendHandshakeTestCase.RunAll(client, logger, 6380)
}
