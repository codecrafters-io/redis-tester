package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/filesystem_asserter"
	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofCreateAofDirectory(stageHarness *test_case_harness.TestCaseHarness) error {
	if err := testAofCreateAofDirectoryOnAppendOnlyYes(stageHarness); err != nil {
		return err
	}

	return testAofDontCreateAofDirectoryOnAppendOnlyNo(stageHarness)
}

func testAofCreateAofDirectoryOnAppendOnlyYes(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof-1")

	if err != nil {
		return err
	}

	baseNames := random.RandomWords(2)
	appendDirNameFlag := baseNames[0]
	appendFileNameFlag := fmt.Sprintf("%s.aof", baseNames[1])

	b := redis_executable.NewRedisExecutable(stageHarness)

	// Ensures that the temporary working directory is deleted AFTER the executable is killed
	stageHarness.RegisterTeardownFunc(func() { os.RemoveAll(dataDirectory) })

	if err := b.Run(
		"--dir", dataDirectory,
		"--appendonly", "yes",
		"--appenddirname", appendDirNameFlag,
		"--appendfilename", appendFileNameFlag,
		"--appendfsync", "always",
	); err != nil {
		return err
	}

	logger := stageHarness.Logger

	fsAsserter := filesystem_asserter.NewFilesystemAsserter([]filesystem_assertion.FilesystemAssertion{
		&filesystem_assertion.DirExistsAssertion{
			AbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag),
		}},
	)

	if err := fsAsserter.RunAssertions(logger); err != nil {
		return err
	}

	if err := b.Kill(); err != nil {
		return err
	}

	return nil
}

func testAofDontCreateAofDirectoryOnAppendOnlyNo(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof-2")

	if err != nil {
		return err
	}

	b := redis_executable.NewRedisExecutable(stageHarness)

	stageHarness.RegisterTeardownFunc(func() { os.RemoveAll(dataDirectory) })

	// Launch with --dir only (appendonly defaults to no)
	if err := b.Run("--dir", dataDirectory); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// Default appenddirname is "appendonlydir"; it must not be created when AOF is off.
	appendOnlyDirPath := filepath.Join(dataDirectory, "appendonlydir")

	// Sleep 100ms because even in the 'error' case where the executable creates the directory
	// the directory might not have been created so early so it might accidentally pass the test
	time.Sleep(100 * time.Millisecond)

	fsAsserter := filesystem_asserter.NewFilesystemAsserter([]filesystem_assertion.FilesystemAssertion{
		&filesystem_assertion.DirDoesNotExistAssertion{
			AbsolutePath: appendOnlyDirPath,
		},
	})

	if err := fsAsserter.RunAssertions(logger); err != nil {
		return err
	}

	return nil
}
