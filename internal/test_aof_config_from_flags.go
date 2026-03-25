package internal

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofConfigFromFlags(stageHarness *test_case_harness.TestCaseHarness) error {
	workingDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	baseNames := testerutils_random.RandomWords(2)
	appendDirName := baseNames[0]
	appendFileName := fmt.Sprintf("%s.aof", baseNames[1])

	b := redis_executable.NewRedisExecutable(stageHarness)

	// Ensures that the temporary working directory is deleted AFTER the executable is killed
	stageHarness.RegisterTeardownFunc(func() { os.RemoveAll(workingDirectory) })

	if err := b.Run(
		"--dir", workingDirectory,
		"--appendonly", "yes",
		"--appenddirname", appendDirName,
		"--appendfilename", appendFileName,
	); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := (&ClientsSpawner{Addr: "localhost:6379", StageHarness: stageHarness}).SpawnClientWithPrefix("client")

	if err != nil {
		return err
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"CONFIG", "GET", "dir"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("dir", workingDirectory),
			},
			{
				Command:   []string{"CONFIG", "GET", "appendonly"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendonly", "yes"),
			},
			{
				Command:   []string{"CONFIG", "GET", "appenddirname"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appenddirname", appendDirName),
			},
			{
				Command:   []string{"CONFIG", "GET", "appendfilename"},
				Assertion: resp_assertions.NewConfigGetBulkStringValueAssertion("appendfilename", appendFileName),
			},
		},
	}

	if err := multiCommandTestCase.RunAll(client, logger); err != nil {
		return err
	}

	return nil
}
