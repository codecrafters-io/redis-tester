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
	key1InitialValue, key2InitialValue := initialValues[0], initialValues[1]
	key1NewValue, key2NewValue := newValues[0], newValues[1]

	// Client 1: Set initial values for both keys
	setKeysTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, strconv.Itoa(key1InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, strconv.Itoa(key2InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}

	if err := setKeysTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Watch both keys
	watchTestCase := test_cases.WatchTestCase{Keys: []string{key1, key2}}

	if err := watchTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Queue a transaction that updates key2
	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{{"SET", key2, strconv.Itoa(key2NewValue)}},
		// Expect nil since a watched key will be modified below
		ExpectedResponseArray: nil,
	}

	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Modify key1 (one of the watched keys): This should error out the following transaction
	modifyKeyTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key1, strconv.Itoa(key1NewValue)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := modifyKeyTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: EXEC should return nil array since a watched key was modified
	if err := transactionTestCase.RunExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Verify key2 still holds its initial value (because transaction returns nil)
	return (&test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key2},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(key2InitialValue)),
	}).Run(clients[1], logger)
}
