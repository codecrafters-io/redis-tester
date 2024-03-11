package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroReplicas(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	// Run the user's code as a master
	masterBinary := redis_executable.NewRedisExecutable(stageHarness)
	if err := masterBinary.Run([]string{
		"--port", "6379",
	}); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	waitTestCase := test_cases.ReplicationTestCase{}

	return waitTestCase.RunWait(client, logger, "0", "60000", 0)
}
