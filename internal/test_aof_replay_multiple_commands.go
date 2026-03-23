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

func testAofReplayMultipleCommands(stageHarness *test_case_harness.TestCaseHarness) error {
	workingDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(5)
	appendDirName := names[0]
	appendFileName := fmt.Sprintf("%s.aof", names[1])
	actualAppendFileName := fmt.Sprintf("%s.aof", names[2])
	key1 := names[3]
	key2 := names[4]
	value1 := strconv.Itoa(random.RandomInt(100, 500))
	value2 := strconv.Itoa(random.RandomInt(100, 500))
	value3 := strconv.Itoa(random.RandomInt(100, 500))

	aofDirectoryCreator := AofDirectoryCreator{
		WorkingDirectory:             workingDirectory,
		AppendDirName:                appendDirName,
		AppendFileNameinFlag:         appendFileName,
		AppendOnlyFilenameInManifest: actualAppendFileName,
		CommandsInsideAppendOnlyFile: [][]string{
			{"SET", key1, value1},
			{"SET", key2, value2},
			{"SET", key1, value3},
		},
	}

	if err := aofDirectoryCreator.Create(logger); err != nil {
		return err
	}

	b := redis_executable.NewRedisExecutable(stageHarness)

	stageHarness.RegisterTeardownFunc(func() { os.RemoveAll(workingDirectory) })

	if err := b.Run(
		"--dir", workingDirectory,
		"--appendonly", "yes",
		"--appenddirname", appendDirName,
		"--appendfilename", appendFileName,
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

	logger.Infof("Checking if commands in append-only file were replayed")

	getKey1 := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key1},
		Assertion: resp_assertions.NewBulkStringAssertion(value3),
	}
	if err := getKey1.Run(client, logger); err != nil {
		return err
	}

	getKey2 := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{key2},
		Assertion: resp_assertions.NewBulkStringAssertion(value2),
	}
	return getKey2.Run(client, logger)
}
