package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-tester/internal/filesystem_asserter"
	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAofCreateAofDirectory(stageHarness *test_case_harness.TestCaseHarness) error {
	dataDirectory, err := MkdirTemp("aof")

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
	); err != nil {
		return err
	}

	logger := stageHarness.Logger

	fsAsserter := filesystem_asserter.NewFilesystemAsserter([]filesystem_assertion.FilesystemAssertion{
		&filesystem_assertion.DirExistsAssertion{
			AbsolutePath: filepath.Join(dataDirectory, appendDirNameFlag),
		}},
	)

	return fsAsserter.RunAssertions(logger)
}
