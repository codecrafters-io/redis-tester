package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxDiscard(stageHarness *test_case_harness.TestCaseHarness) error {
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

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{"bar", "42"},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", "foo", "41"},
			{"INCR", "foo"},
		},
		ResultArray: []resp_value.Value{},
	}

	if err := transactionTestCase.RunWithoutExec(client, logger); err != nil {
		return err
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"DISCARD"},
			{"GET", "foo"},
			{"GET", "bar"},
			{"DISCARD"},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("OK"),
			resp_assertions.NewNilAssertion(),
			resp_assertions.NewStringAssertion("42"),
			resp_assertions.NewErrorAssertion("ERR DISCARD without MULTI"),
		},
	}

	if err := multiCommandTestCase.RunAll(client, logger); err != nil {
		return err
	}

	return nil
}
