package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingTrackingKeyModification(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keys := testerutils_random.RandomWords(4)
	watchedKey, unwatchedKey := keys[0], keys[1]
	watcherClient, modifierClient := clients[0], clients[1]

	logger.Infof("Testing optimistic locking when watched key is modified")

	// Transaction failing case
	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:              watcherClient,
		ModifierClient:             modifierClient,
		InitialKeys:                []string{watchedKey, unwatchedKey},
		KeysWatchedByWatcherClient: []string{watchedKey},
		KeyToBeModifiedByWatcherClientInTransaction: unwatchedKey,
		KeyToBeModifiedByModifierClient:             watchedKey,
	}

	if err := optimisticLockingTestCase.Run(logger); err != nil {
		return err
	}

	logger.Infof("Testing optimistic locking when watched key is not modified")

	// Transaction succeeding case
	newClients, err := clientsSpawner.SpawnClients(2)

	if err != nil {
		return err
	}

	watcherClient, modifierClient = newClients[0], newClients[1]
	watchedKey, unwatchedKey = keys[2], keys[3]
	optimisticLockingTestCase2 := test_cases.OptimisticLockingTestCase{
		WatcherClient:                               watcherClient,
		ModifierClient:                              modifierClient,
		InitialKeys:                                 []string{watchedKey, unwatchedKey},
		KeysWatchedByWatcherClient:                  []string{watchedKey},
		KeyToBeModifiedByModifierClient:             unwatchedKey,
		KeyToBeModifiedByWatcherClientInTransaction: unwatchedKey,
	}

	return optimisticLockingTestCase2.Run(logger)
}
