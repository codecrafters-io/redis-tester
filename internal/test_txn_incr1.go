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

func testTxIncr1(stageHarness *test_case_harness.TestCaseHarness) error {
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

	value := random.RandomInt(1, 100)
	key := random.RandomWord()

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key, fmt.Sprint(value)},
				Assertion: resp_assertions.NewStringAssertion("OK"),
			},
			{
				Command:   []string{"INCR", key},
				Assertion: resp_assertions.NewIntegerAssertion(value + 1),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
