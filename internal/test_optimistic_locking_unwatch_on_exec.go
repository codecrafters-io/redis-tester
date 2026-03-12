package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
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

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
		Logger:       logger,
	}
	clients, err := clientsSpawner.SpawnClients(2)
	if err != nil {
		return err
	}

	watcherClient, modifierClient := clients[0], clients[1]

	keys := testerutils_random.RandomWords(2)
	key1, key2 := keys[0], keys[1]

	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:                               watcherClient,
		ModifierClient:                              modifierClient,
		InitialKeys:                                 keys,
		KeysWatchedByWatcherClient:                  []string{key1},
		KeyToBeModifiedByModifierClient:             key1,
		KeyToBeModifiedByWatcherClientInTransaction: key2,
	}

	// First run: modifier client modifies a watched key, transaction fails
	if err := optimisticLockingTestCase.Run(logger); err != nil {
		return err
	}

	// Reset watched keys becuse the test case's transaction EXEC will clear the watched keys
	optimisticLockingTestCase.ResetWatchedKeys()

	// Retry the same transaction: Should pass -> Because previous transaction's EXEC clears watched keys
	if err := optimisticLockingTestCase.RunTransaction(logger); err != nil {
		return err
	}

	// Check if the new value (set by the transaction persists)
	return optimisticLockingTestCase.RunValueCheckOfKeyModifiedInTransaction(logger)
}
