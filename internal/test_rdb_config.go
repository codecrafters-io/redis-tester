package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbConfig(stageHarness *test_case_harness.TestCaseHarness) error {
	tmpDir, err := os.MkdirTemp("", "rdbfiles")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// On MacOS, the tmpDir is a symlink to a directory in /var/folders/...
	realPath, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return fmt.Errorf("CodeCrafters tester error: could not resolve symlink: %v", err)
	}
	tmpDir = realPath

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run("--dir", tmpDir,
		"--dbfilename", fmt.Sprintf("%s.rdb", testerutils_random.RandomWord())); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "")
	if err != nil {
		return err
	}

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "config",
		Args:      []string{"get", "dir"},
		Assertion: resp_assertions.NewStringArrayAssertion([]string{"dir", tmpDir}),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()
	return nil
}
