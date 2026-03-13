package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
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

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	clients, err := clientsSpawner.SpawnClients(2)
	if err != nil {
		return err
	}

	keys := testerutils_random.RandomWords(2)
	key1, key2 := keys[0], keys[1]

	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:                               clients[0],
		ModifierClient:                              clients[1],
		InitialKeys:                                 []string{key1, key2},
		KeysWatchedByWatcherClient:                  []string{key1, key2},
		KeyToBeModifiedByModifierClient:             key1,
		KeyToBeModifiedByWatcherClientInTransaction: key1,
	}

	if err := optimisticLockingTestCase.RunSetInitialKeys(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunWatchKeys(logger); err != nil {
		return err
	}

	// Modify key1: This should have failed the upcoming transaction because a watched key (key1)
	// is modified, but this does not happen because the following UNWATCH clears the watched keys
	if err := optimisticLockingTestCase.RunSetKeyUsingModifierClient(logger); err != nil {
		return err
	}

	if err := optimisticLockingTestCase.RunUnwatch(logger); err != nil {
		return err
	}

	// Clear the watched keys after unwatch is run
	optimisticLockingTestCase.ResetWatchedKeys()

	// Run the transaction that modifies key1 (previously watched key; but now has been unwatched)
	if err := optimisticLockingTestCase.RunTransaction(logger); err != nil {
		return err
	}

	// The new value of key2 should overwrite the old one because the transaction above should succeed
	return optimisticLockingTestCase.RunValueCheckOfKeyModifiedInTransaction(logger)
}
