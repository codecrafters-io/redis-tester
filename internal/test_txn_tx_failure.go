package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxErr(stageHarness *test_case_harness.TestCaseHarness) error {
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
	key1, key2 := uniqueKeys[0], uniqueKeys[1]
	value1 := uniqueKeys[2]
	value2 := random.RandomInt(1, 100)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, value1},
				Assertion: resp_assertions.NewStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, fmt.Sprint(value2)},
				Assertion: resp_assertions.NewStringAssertion("OK"),
			},
		},
	}

	if err := multiCommandTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"INCR", key1},
			{"INCR", key2},
		},
		ExpectedResponseArray: []resp_assertions.RESPAssertion{
			resp_assertions.NewErrorAssertion("ERR value is not an integer or out of range"),
			resp_assertions.NewIntegerAssertion(value2 + 1),
		},
	}

	if err := transactionTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	multiCommandTestCase = test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"GET", key2},
				Assertion: resp_assertions.NewStringAssertion(fmt.Sprint(value2 + 1)),
			},
			{
				Command:   []string{"GET", key1},
				Assertion: resp_assertions.NewStringAssertion(value1),
			},
		},
	}

	return multiCommandTestCase.RunAll(clients[1], logger)
}
