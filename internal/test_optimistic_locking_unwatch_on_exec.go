package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingUnwatchOnExec(stageHarness *test_case_harness.TestCaseHarness) error {
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
	newValues := testerutils_random.RandomInts(200, 500, 1)
	finalValues := testerutils_random.RandomInts(1000, 2000, 2)

	key1, key2 := keys[0], keys[1]
	newValue1 := newValues[0]
	finalValue1, finalValue2 := finalValues[0], finalValues[1]

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

	// Client 1: First transaction: This will fail because the watched key was modified
	abortedTxn := test_cases.TransactionTestCase{
		CommandQueue:          [][]string{{"SET", key2, strconv.Itoa(finalValue2)}},
		ExpectedResponseArray: nil,
	}
	if err := abortedTxn.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Second transaction: This will succeed because the previous exec will have cleared the watched keys
	successTxn := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", key1, strconv.Itoa(finalValue1)},
			{"SET", key2, strconv.Itoa(finalValue2)},
		},
		ExpectedResponseArray: []resp_assertions.RESPAssertion{
			resp_assertions.NewSimpleStringAssertion("OK"),
			resp_assertions.NewSimpleStringAssertion("OK"),
		},
	}
	if err := successTxn.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Verify the previous transaction suceeded
	return (&test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"GET", key1},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(finalValue1)),
			},
			{
				Command:   []string{"GET", key2},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(finalValue2)),
			},
		},
	}).RunAll(clients[1], logger)
}
