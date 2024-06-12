package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxSuccess(stageHarness *test_case_harness.TestCaseHarness) error {
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
		Command:   "MULTI",
		Args:      []string{},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{"foo", "6"},
		Assertion: resp_assertions.NewStringAssertion("QUEUED"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "INCR",
		Args:      []string{"foo"},
		Assertion: resp_assertions.NewStringAssertion("QUEUED"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "INCR",
		Args:      []string{"bar"},
		Assertion: resp_assertions.NewStringAssertion("QUEUED"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{"bar"},
		Assertion: resp_assertions.NewStringAssertion("QUEUED"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	setCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "EXEC",
		Args:      []string{},
		Assertion: resp_assertions.NewOrderedArrayAssertion([]resp_value.Value{resp_value.NewSimpleStringValue("OK"), resp_value.NewIntegerValue(7), resp_value.NewIntegerValue(1), resp_value.NewBulkStringValue("1")}),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
