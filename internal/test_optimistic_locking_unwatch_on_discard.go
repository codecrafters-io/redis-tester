package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingUnwatchOnDiscard(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
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
		KeysWatchedByWatcherClient:                  []string{key1, key2},
		KeyToBeModifiedByModifierClient:             key1,
		KeyToBeModifiedByWatcherClientInTransaction: key2,
	}

	if err := optimisticLockingTestCase.RunSetInitialKeys(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunWatchKeys(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunTransactionWithoutExec(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunSetKeyUsingModifierClient(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunDiscardTransaction(logger); err != nil {
		return err
	}

	// The previous DISCARD clears all watched keys
	optimisticLockingTestCase.ResetWatchedKeys()

	// Retry the transaction again: Should succeed since the watched keys were reset by DISCARD
	if err := optimisticLockingTestCase.RunTransaction(logger); err != nil {
		return err
	}

	// Check the value of modified key in the transaction
	return optimisticLockingTestCase.RunValueCheckOfKeyModifiedInTransaction(logger)
}
