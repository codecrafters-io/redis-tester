package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofWriteMultipleCommands(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(4)
	appendDirNameFlag := names[0]
	appendFileNameFlag := fmt.Sprintf("%s.aof", names[1])
	appendFileBaseName := fmt.Sprintf("%s.1.incr.aof", appendFileNameFlag)

	b := redis_executable.NewRedisExecutable(stageHarness)

	// Remove the working directory after redis has quit
	stageHarness.RegisterTeardownFunc(func() {
		os.RemoveAll(dataDirectory)
	})

	if err := b.Run(
		"--dir", dataDirectory,
		"--appendonly", "yes",
		"--appenddirname", appendDirNameFlag,
		"--appendfilename", appendFileNameFlag,
		"--appendfsync", "always",
	); err != nil {
		return err
	}

	client, err := (&ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}).SpawnClientWithPrefix("client")

	if err != nil {
		return err
	}

	listKey := names[2]
	listElement := names[3]

	lPushCmd := []string{"LPUSH", listKey, listElement}
	blpopCmd := []string{"BLPOP", listKey, "0"}

	aofWriteTestCase := test_cases.AofWriteTestCase{
		AppendOnlyFileAbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag, appendFileBaseName),
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   lPushCmd,
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			{
				Command:   blpopCmd,
				Assertion: resp_assertions.NewOrderedBulkStringArrayAssertion([]string{listKey, listElement}),
			},
		},
		ExpectedCommandsInAppendOnlyFile: [][]string{
			lPushCmd,
			// BLPOP is transformed to LPOP
			{"LPOP", listKey},
		},
	}

	return aofWriteTestCase.Run(client, logger)
}
