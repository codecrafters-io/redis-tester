package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
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
	for _, c := range clients {
		defer c.Close()
	}

	keys := testerutils_random.RandomWords(2)
	initialValues := testerutils_random.RandomInts(1, 100, 2)
	newValues := testerutils_random.RandomInts(200, 500, 3)

	key1, key2 := keys[0], keys[1]
	newValue1, newValue2, newValue3 := newValues[0], newValues[1], newValues[2]

	// Client 1: Set initial values for both keys
	if err := (&test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, strconv.Itoa(initialValues[0])},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, strconv.Itoa(initialValues[1])},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}).RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Watch both keys
	if err := (test_cases.WatchTestCase{Keys: []string{key1, key2}}).Run(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Modify key1 (a watched key)
	if err := (&test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key1, strconv.Itoa(newValue1)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}).Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: UNWATCH to clear all watched keys
	if err := (&test_cases.SendCommandTestCase{
		Command:   "UNWATCH",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}).Run(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Queue a transaction updating both keys
	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", key1, strconv.Itoa(newValue3)},
			{"SET", key2, strconv.Itoa(newValue2)},
		},
		// Transaction should succeed since UNWATCH was issued from client 0 before
		ExpectedResponseArray: []resp_assertions.RESPAssertion{
			resp_assertions.NewSimpleStringAssertion("OK"),
			resp_assertions.NewSimpleStringAssertion("OK"),
		},
	}
	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	if err := transactionTestCase.RunExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Verify that the transaction was executed
	return (&test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"GET", key1},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(newValue3)),
			},
			{
				Command:   []string{"GET", key2},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(newValue2)),
			},
		},
	}).RunAll(clients[1], logger)
}
