package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
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

	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:  clients[0],
		ModifierClient: clients[1],
		InitialKeys:    keys,
		// Watch all keys
		KeysWatchedByWatcherClient: keys,
		// Modify any of the watched keys
		KeyToBeModifiedByModifierClient: random.RandomElementFromArray(keys),
		// Use transaction to modify any of the key: Should fail in any case
		KeyToBeModifiedByWatcherClientInTransaction: random.RandomElementFromArray(keys),
	}

	return optimisticLockingTestCase.Run(logger)
}
