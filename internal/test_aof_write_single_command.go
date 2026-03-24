package internal

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofWriteSingleCommand(stageHarness *test_case_harness.TestCaseHarness) error {
	workingDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(3)
	appendDirName := names[0]
	appendFileName := fmt.Sprintf("%s.aof", names[1])
	actualAppendFileName := fmt.Sprintf("%s.aof", names[2])
	key := random.RandomWord()
	value := strconv.Itoa(random.RandomInt(100, 500))

	aofDirectoryCreator := AofDirectoryCreator{
		WorkingDirectory:             workingDirectory,
		AppendDirName:                appendDirName,
		AppendFileNameinFlag:         appendFileName,
		AppendOnlyFilenameInManifest: actualAppendFileName,
		CommandsInsideAppendOnlyFile: [][]string{},
	}

	if err := aofDirectoryCreator.Create(logger); err != nil {
		return err
	}

	b := redis_executable.NewRedisExecutable(stageHarness)

	stageHarness.RegisterTeardownFunc(func() {
		aofDirectoryCreator.Cleanup(stageHarness)
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

	setCommand := []string{"SET", key, value}
	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   setCommand[0],
		Args:      setCommand[1:],
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	filesystemAsserter := filesystem_assertion.NewFileSystemAsserter(
		[]filesystem_assertion.FilesystemAssertion{
			filesystem_assertion.AofAppendOnlyFileAssertion{
				AbsolutePath:     filepath.Join(workingDirectory, appendDirName, fmt.Sprintf("%s.1.incr.aof", actualAppendFileName)),
				ExpectedCommands: []string{strings.Join(setCommand, " ")},
			},
		},
	)

	return filesystemAsserter.RunAssertions(logger)
}
