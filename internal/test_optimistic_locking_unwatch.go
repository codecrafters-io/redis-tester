package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingUnwatch(stageHarness *test_case_harness.TestCaseHarness) error {
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

	uniqueKey := random.RandomWord()
	value := random.RandomInt(1, 100)

	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{uniqueKey, fmt.Sprint(value)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	watchCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "WATCH",
		Args:      []string{uniqueKey},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := watchCommandTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	unwatchCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "UNWATCH",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := unwatchCommandTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	modifyCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{uniqueKey, fmt.Sprint(value + 10)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := modifyCommandTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"INCR", uniqueKey},
		},
		ExpectedResponseArray: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(value + 11),
		},
	}

	return transactionTestCase.RunAll(clients[0], logger)
}
