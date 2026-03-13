package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroReplicas(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run("--port", "6379"); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client, err := clientsSpawner.SpawnClientWithPrefix("client")
	if err != nil {
		return err
	}

	waitTestCase := test_cases.WaitTestCase{
		Replicas:              0,
		TimeoutInMilliseconds: 60000,
		ExpectedMessage:       0,
	}

	return waitTestCase.Run(client, logger)
}
