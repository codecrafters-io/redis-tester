package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofAppendfsyncFromFlags(stageHarness *test_case_harness.TestCaseHarness) error {
	appendfsyncValue := "always"

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run("--appendfsync", appendfsyncValue); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := (&ClientsSpawner{Addr: "localhost:6379", StageHarness: stageHarness}).SpawnClientWithPrefix("client")

	if err != nil {
		return err
	}

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "CONFIG",
		Args:      []string{"GET", "appendfsync"},
		Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendfsync", appendfsyncValue),
	}

	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
