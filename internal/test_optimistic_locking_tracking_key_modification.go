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
	clients, err := SpawnClients(4, "localhost:6379", stageHarness, logger)

	if err != nil {
		return err
	}

	for _, c := range clients {
		defer c.Close()
	}

	keys := testerutils_random.RandomWords(4)
	watchedKey, unwatchedKey := keys[0], keys[1]

	logger.Infof("Testing optimistic locking when watched key is modified")

	// Transaction failing case
	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:              clients[0],
		ModifierClient:             clients[1],
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
	watchedKey, unwatchedKey = keys[2], keys[3]
	optimisticLockingTestCase2 := test_cases.OptimisticLockingTestCase{
		WatcherClient:                               clients[2],
		ModifierClient:                              clients[3],
		InitialKeys:                                 []string{watchedKey, unwatchedKey},
		KeysWatchedByWatcherClient:                  []string{watchedKey},
		KeyToBeModifiedByModifierClient:             unwatchedKey,
		KeyToBeModifiedByWatcherClientInTransaction: unwatchedKey,
	}

	return optimisticLockingTestCase2.Run(logger)
}
