package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxEmpty(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	emptyTransactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{},
		ResultArray:  []resp_value.Value{},
	}

	if err := emptyTransactionTestCase.RunAll(client, logger); err != nil {
		return err
	}

	bareExecCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "EXEC",
		Args:      []string{},
		Assertion: resp_assertions.NewErrorAssertion("ERR EXEC without MULTI"),
	}

	return bareExecCommandTestCase.Run(client, logger)
}
