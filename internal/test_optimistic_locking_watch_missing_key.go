package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingWatchMissingKey(stageHarness *test_case_harness.TestCaseHarness) error {
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

	key := testerutils_random.RandomWord()

	optimisticLockingTestCase := test_cases.OptimisticLockingTestCase{
		WatcherClient:  clients[0],
		ModifierClient: clients[1],
		// Set no keys initially
		InitialKeys:                                 []string{},
		KeysWatchedByWatcherClient:                  []string{key},
		KeyToBeModifiedByModifierClient:             key,
		KeyToBeModifiedByWatcherClientInTransaction: key,
	}

	return optimisticLockingTestCase.Run(logger)
}
