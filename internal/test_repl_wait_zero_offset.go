package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroOffset(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run("--port", "6379"); err != nil {
		return err
	}

	logger := stageHarness.Logger
	defer logger.ResetSecondaryPrefixes()

	replicaCount := testerutils_random.RandomInt(3, 9)
	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	replicas, err := SpawnReplicas(replicaCount, stageHarness, logger, "localhost:6379")
	if err != nil {
		return err
	}
	for _, replica := range replicas {
		defer replica.Close()
	}

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	logger.UpdateLastSecondaryPrefix("test")
	client.UpdateBaseLogger(logger)

	diff := ((replicaCount + 3) - 3) / 3
	safeDiff := max(1, diff) // If diff is 0, it will get stuck in an infinite loop
	for actual := 3; actual < replicaCount+3; actual += safeDiff {
		waitTestCase := test_cases.WaitTestCase{
			Replicas:              actual,
			TimeoutInMilliseconds: 500,
			ExpectedMessage:       replicaCount,
		}
		if err := waitTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
