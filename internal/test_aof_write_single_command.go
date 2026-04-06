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

func testAofWriteSingleCommand(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
	names := random.RandomWords(3)
	appendDirNameFlag := names[0]
	appendFileNameFlag := fmt.Sprintf("%s.aof", names[1])
	appendFileBaseName := fmt.Sprintf("%s.1.incr.aof", appendFileNameFlag)

	key := names[2]
	value := strconv.Itoa(random.RandomInt(100, 500))

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

	setCommand := []string{"SET", key, value}

	aofWriteTestCase := test_cases.AofWriteTestCase{
		AppendOnlyFileAbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag, appendFileBaseName),
		CommandWithAssertions: []test_cases.CommandWithAssertion{{
			Command:   setCommand,
			Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
		}},
	}

	return aofWriteTestCase.Run(client, logger)
}
