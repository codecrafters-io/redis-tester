package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingTrackingKeyModification(stageHarness *test_case_harness.TestCaseHarness) error {
	if err := testOptimisticLockingScenario(stageHarness, true); err != nil {
		return err
	}

	stageHarness.Logger.Infof("Tearing down Redis executable and clients")

	return testOptimisticLockingScenario(stageHarness, false)
}

// testOptimisticLockingScenario runs a WATCH/MULTI/EXEC scenario.
// When modifyWatchedKey is true, client 2 modifies the watched key, causing
// the transaction to abort (EXEC returns nil). Otherwise, client 2 modifies
// the unwatched key and the transaction succeeds.
func testOptimisticLockingScenario(stageHarness *test_case_harness.TestCaseHarness, modifyWatchedKey bool) error {
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
	initialValue1, initialValue2 := initialValues[0], initialValues[1]
	newValue1, newValue2 := newValues[0], newValues[1]

	// Client 1: Set initial values
	setVariablesTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, strconv.Itoa(initialValue1)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, strconv.Itoa(initialValue2)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}
	if err := setVariablesTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Watch key1
	if err := (test_cases.WatchTestCase{Keys: []string{key1}}).Run(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Queue a transaction that updates key2
	var expectedResponseArray []resp_assertions.RESPAssertion
	if !modifyWatchedKey {
		expectedResponseArray = []resp_assertions.RESPAssertion{
			resp_assertions.NewSimpleStringAssertion("OK"),
		}
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue:          [][]string{{"SET", key2, strconv.Itoa(newValue2)}},
		ExpectedResponseArray: expectedResponseArray,
	}
	if err := transactionTestCase.RunWithoutExec(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Modify either the watched or unwatched key
	client2Key, client2Value := key2, newValue2
	if modifyWatchedKey {
		client2Key, client2Value = key1, newValue1
	}

	modifyKeyTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{client2Key, strconv.Itoa(client2Value)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}
	if err := modifyKeyTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: EXEC — aborts if watched key was modified, succeeds otherwise
	if err := transactionTestCase.RunExec(clients[0], logger); err != nil {
		return err
	}

	// On success, verify key2 has the value set by the transaction
	key2ExpectedValue := initialValue2
	if !modifyWatchedKey {
		key2ExpectedValue = newValue2
	}

	getTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key2},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(key2ExpectedValue)),
	}
	return getTestCase.Run(clients[0], logger)

}
