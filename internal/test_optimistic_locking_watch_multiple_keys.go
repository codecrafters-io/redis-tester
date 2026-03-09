package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
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
	for _, c := range clients {
		defer c.Close()
	}

	keys := testerutils_random.RandomWords(2)
	initialValues := testerutils_random.RandomInts(1, 100, 2)
	newValues := testerutils_random.RandomInts(500, 1000, 2)

	key1, key2 := keys[0], keys[1]
	initialValue2 := initialValues[1]
	newValue1, newValue2 := newValues[0], newValues[1]

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

	// Client 1: Queue a transaction that updates key2
	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{{"SET", key2, strconv.Itoa(newValue2)}},
		// Expect abort since a watched key will be modified
		ExpectedResponseArray: nil,
	}
	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Modify key1 (one of the watched keys) — this should abort the transaction
	if err := (&test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key1, strconv.Itoa(newValue1)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}).Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: EXEC — should return nil array since a watched key was modified
	if err := transactionTestCase.RunExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Verify key2 still holds its initial value (transaction was aborted)
	return (&test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key2},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(initialValue2)),
	}).Run(clients[1], logger)
}
