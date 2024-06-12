package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxErr(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client1")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{"foo", "abc"},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{"bar", "7"},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"INCR", "foo"},
			{"INCR", "bar"},
		},
		ResultArray: []resp_value.Value{
			resp_value.NewErrorValue("ERR value is not an integer or out of range"), resp_value.NewIntegerValue(8)},
	}

	return transactionTestCase.RunAll(client, logger)
}
