package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxDiscard(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	uniqueKeys := random.RandomWords(2)
	key1, key2 := uniqueKeys[0], uniqueKeys[1]
	value1, value2 := random.RandomInt(1, 100), random.RandomInt(1, 100)

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key2, fmt.Sprint(value2)},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", key1, fmt.Sprint(value1)},
			{"INCR", key1},
		},
	}

	if err := transactionTestCase.RunWithoutExec(client, logger); err != nil {
		return err
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"DISCARD"},
				Assertion: resp_assertions.NewStringAssertion("OK"),
			},
			{
				Command:   []string{"GET", key1},
				Assertion: resp_assertions.NewNilAssertion(),
			},
			{
				Command:   []string{"GET", key2},
				Assertion: resp_assertions.NewStringAssertion(fmt.Sprint(value2)),
			},
			{
				Command:   []string{"DISCARD"},
				Assertion: resp_assertions.NewErrorAssertion("ERR DISCARD without MULTI"),
			},
		},
	}

	if err := multiCommandTestCase.RunAll(client, logger); err != nil {
		return err
	}

	return nil
}
