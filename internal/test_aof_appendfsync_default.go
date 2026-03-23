package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofAppendfsyncDefault(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)

	if err := b.Run(); err != nil {
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
		Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendfsync", "everysec"),
	}

	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
