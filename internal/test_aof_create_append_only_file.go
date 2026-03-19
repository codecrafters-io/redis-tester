package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofCreateAppendOnlyFile(stageHarness *test_case_harness.TestCaseHarness) error {
	workingDirectory, err := MkdirTemp("aof")

	if err != nil {
		return err
	}

	logger := stageHarness.Logger
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

	appendOnlyFileBaseName := fmt.Sprintf("%s.1.incr.aof", appendFileName)

	filesystemAsserter := filesystem_assertion.NewFileSystemAsserter([]filesystem_assertion.FilesystemAssertion{
		filesystem_assertion.DirExistsAssertion{
			AbsolutePath: filepath.Join(workingDirectory, appendDirName),
		},
		filesystem_assertion.AofAppendOnlyFileAssertion{
			AbsolutePath: filepath.Join(workingDirectory, appendDirName, appendOnlyFileBaseName),
			// Expect no commands to be present in the append-only file
			ExpectedCommands: []string{},
		},
	})

	return filesystemAsserter.RunAssertions(logger)

}
