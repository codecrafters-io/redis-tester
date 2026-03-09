package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingWatchMultipleKeys(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, client := range clients {
		defer client.Close()
	}

	uniqueKeys := random.RandomWords(3)
	key1, key2, key3 := uniqueKeys[0], uniqueKeys[1], uniqueKeys[2]
	value := random.RandomInt(1, 100)

	setCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, fmt.Sprint(value)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, fmt.Sprint(value)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key3, fmt.Sprint(value)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}

	if err := setCommandTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	watchCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "WATCH",
		Args:      []string{key1, key2, key3},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := watchCommandTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"INCR", key1},
		},
	}

	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	modifyCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key2, fmt.Sprint(value + 10)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := modifyCommandTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	execCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "EXEC",
		Args:      []string{},
		Assertion: resp_assertions.NewNilArrayAssertion(),
	}

	return execCommandTestCase.Run(clients[0], logger)
}
