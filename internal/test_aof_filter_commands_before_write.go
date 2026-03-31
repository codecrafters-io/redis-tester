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

func testAofFilterCommandsBeforeWrite(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(7)
	appendDirNameFlag := names[0]
	appendFileNameFlag := fmt.Sprintf("%s.aof", names[1])
	appendFileBaseName := fmt.Sprintf("%s.1.incr.aof", appendFileNameFlag)
	key1, value1 := names[2], names[3]
	echoArg := names[4]
	key2, value2 := names[5], names[6]

	b := redis_executable.NewRedisExecutable(stageHarness)

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

	setCommand1 := []string{"SET", key1, value1}
	getCommand := []string{"GET", key1}
	echoCommand := []string{"ECHO", echoArg}
	setCommand2 := []string{"SET", key2, value2}

	aofWriteTestCase := test_cases.AofWriteTestCase{
		AppendOnlyFileAbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag, appendFileBaseName),
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   setCommand1,
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   getCommand,
				Assertion: resp_assertions.NewBulkStringAssertion(value1),
			},
			{
				Command:   echoCommand,
				Assertion: resp_assertions.NewBulkStringAssertion(echoArg),
			},
			{
				Command:   setCommand2,
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
		ExpectedCommandsInAppendOnlyFile: [][]string{
			setCommand1,
			setCommand2,
		},
	}

	return aofWriteTestCase.Run(client, logger)
}
