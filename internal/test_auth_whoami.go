package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWhoami(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)

	if err := b.Run(); err != nil {
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

	whoamiTestCase := test_cases.AclWhoamiTestCase{
		ExpectedUsername: "default",
	}

	return whoamiTestCase.Run(client, logger)
}
