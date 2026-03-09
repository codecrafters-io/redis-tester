package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingWatchMissingKeys(stageHarness *test_case_harness.TestCaseHarness) error {
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

	key := testerutils_random.RandomWords(1)[0]
	newValue1 := testerutils_random.RandomInt(1, 100)
	newValue2 := testerutils_random.RandomInt(500, 1000)

	// Client 1: Watch a key that doesn't exist yet
	if err := (test_cases.WatchTestCase{Keys: []string{key}}).Run(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Queue a transaction that updates the watched key
	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue:          [][]string{{"SET", key, strconv.Itoa(newValue2)}},
		ExpectedResponseArray: nil, // Expect nil array since the watched key was modified
	}
	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Set the watched key (creating it): This should fail the transaction
	if err := (&test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key, strconv.Itoa(newValue1)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}).Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: EXEC should return nil array since the watched key was created/modified
	if err := transactionTestCase.RunExec(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Verify that the transaction was aborted
	return (&test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(newValue1)),
	}).Run(clients[0], logger)
}
