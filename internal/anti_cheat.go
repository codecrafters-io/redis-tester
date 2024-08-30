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
	logger := stageHarness.Logger

	b := redis_executable.NewRedisExecutable(stageHarness)
	// If we can't run the executable, it must be an internal error.
	if err := b.Run(); err != nil {
		logger.Criticalf("CodeCrafters internal error. Error instantiating executable: %v", err)
		logger.Criticalf("Try again? Please contact us at hello@codecrafters.io if this persists.")
		return fmt.Errorf("anti-cheat (ac1) failed")
	}

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "replica")
	// If we are unable to connect to the redis server, it is okay to skip anti-cheat in that case, their server must not be working.
	if err != nil {
		return nil
	}
	defer client.Close()

	// All the answers for MEMORY DOCTOR include the string "sam" in them.
	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "MEMORY",
		Args:                      []string{"DOCTOR"},
		Assertion:                 resp_assertions.NewRegexStringAssertion("[sS]am"),
		ShouldSkipUnreadDataCheck: true,
	}
	err = commandTestCase.Run(client, logger)

	if err == nil {
		logger.Criticalf("anti-cheat (ac1) failed.")
		logger.Criticalf("Please contact us at hello@codecrafters.io if you think this is a mistake.")
		return fmt.Errorf("anti-cheat (ac1) failed")
	} else {
		return nil
	}
}
