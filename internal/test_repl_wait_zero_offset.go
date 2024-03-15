package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroOffset(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run([]string{
		"--port", "6379",
	}...); err != nil {
		return err
	}

	logger := stageHarness.Logger

	replicaCount := testerutils_random.RandomInt(3, 9)
	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	replicas, err := SpawnReplicas(replicaCount, stageHarness, logger, "localhost:6379")
	if err != nil {
		return err
	}
	for _, replica := range replicas {
		defer replica.Close()
	}

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	waitTestCase := test_cases.ReplicationTestCase{}

	diff := ((replicaCount + 3) - 3) / 3
	safeDiff := max(1, diff) // If diff is 0, it will get stuck in an infinite loop
	for i := 3; i < replicaCount+3; i += safeDiff {
		actual, expected := strconv.Itoa(i), replicaCount
		if err := waitTestCase.RunWait(client, logger, actual, "500", expected); err != nil {
			return err
		}
	}

	return nil
}
