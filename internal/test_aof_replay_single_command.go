package internal

import (
	"fmt"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofReplaySingleCommand(stageHarness *test_case_harness.TestCaseHarness) error {
	workingDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(4)
	appendDirName := names[0]
	appendFileName := fmt.Sprintf("%s.aof", names[1])
	actualAppendFileName := fmt.Sprintf("%s.aof", names[2])
	key := names[3]
	value := strconv.Itoa(random.RandomInt(100, 500))

	aofDirectoryCreator := AofDirectoryCreator{
		DataDirectory:                workingDirectory,
		AppendDirName:                appendDirName,
		AppendFileNameInFlag:         appendFileName,
		AppendOnlyFileNameInManifest: actualAppendFileName,
		CommandsInsideAppendOnlyFile: [][]string{
			{"SET", key, value},
		},
	}

	if err := aofDirectoryCreator.Create(logger); err != nil {
		return err
	}

	b := redis_executable.NewRedisExecutable(stageHarness)

	stageHarness.RegisterTeardownFunc(func() {
		os.RemoveAll(workingDirectory)
	})

	if err := b.Run(
		"--dir", workingDirectory,
		"--appendonly", "yes",
		"--appenddirname", appendDirName,
		"--appendfilename", appendFileName,
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

	logger.Infof("Checking if the command in append-only file was replayed")

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key},
		Assertion: resp_assertions.NewBulkStringAssertion(value),
	}

	return sendCommandTestCase.Run(client, logger)
}
