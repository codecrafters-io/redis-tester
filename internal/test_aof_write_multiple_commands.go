package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

	key1 := names[2]
	key2 := names[3]
	values := random.RandomInts(100, 500, 2)
	value1 := strconv.Itoa(values[0])
	value2 := strconv.Itoa(values[0])

	setCommand1 := []string{"SET", key1, value1}
	setCommand2 := []string{"SET", key2, value2}

	aofWriteTestCase := test_cases.AofWriteTestCase{
		AppendOnlyFileAbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag, appendFileBaseName),
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   setCommand1,
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   setCommand2,
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
		ExpectedCommandsInAppendOnlyFile: [][]string{setCommand1, setCommand2},
	}

	return aofWriteTestCase.Run(client, logger)
}
