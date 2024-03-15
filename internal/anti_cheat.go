package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func antiCheatTest(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "replica")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	// All the answers for MEMORY DOCTOR include the string "sam" in them.
	commandTestCase := test_cases.SendCommandAndReceiveValueTestCase{
		Command:                   "MEMORY",
		Args:                      []string{"DOCTOR"},
		Assertion:                 resp_assertions.NewRegexStringAssertion("[sS]am"),
		ShouldSkipUnreadDataCheck: true,
	}
	err = commandTestCase.Run(client, logger)

	if err == nil {
		logger.Criticalf("anti-cheat (ac1) failed.")
		logger.Criticalf("Are you sure you aren't running this against the actual Redis?")
		return fmt.Errorf("anti-cheat (ac1) failed")
	} else {
		return nil
	}
}
