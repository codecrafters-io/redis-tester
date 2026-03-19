package internal

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofConfigDefaults(stageHarness *test_case_harness.TestCaseHarness) error {
	// On MacOS, the tmpDir is a symlink to a directory in /var/folders/...
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := (&ClientsSpawner{Addr: "localhost:6379", StageHarness: stageHarness}).SpawnClientWithPrefix("client")

	if err != nil {
		return err
	}

	currentWorkingDirectory, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("Error retrieving working directory: %s", err)
	}

	// Test the default configs used for AOF
	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"CONFIG", "GET", "dir"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("dir", currentWorkingDirectory),
			},
			{
				Command:   []string{"CONFIG", "GET", "appendonly"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendonly", "no"),
			},
			{
				Command:   []string{"CONFIG", "GET", "appenddirname"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appenddirname", "appendonlydir"),
			},
			{
				Command:   []string{"CONFIG", "GET", "appendfilename"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendfilename", "appendonly.aof"),
			},
		},
	}

	if err := multiCommandTestCase.RunAll(client, logger); err != nil {
		return err
	}

	return nil

}
