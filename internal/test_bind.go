package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBindToPort(stageHarness *test_case_harness.TestCaseHarness) error {
	port := 6379

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	logger.Infof("Connecting to port %d...", port)

	bindTestCase := test_cases.BindTestCase{
		Port:    port,
		Retries: 15,
	}

	return bindTestCase.Run(b, logger)
}
