package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

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
	client := NewRedisClient("localhost:6379")

	logger.Infof("$ redis-cli CONFIG GET dir")
	resp, err := client.ConfigGet("dir").Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if len(resp) != 2 {
		return fmt.Errorf("Expected 2 elements in response, got %d", len(resp))
	}

	if resp[0] != "dir" {
		return fmt.Errorf("Expected first element in response to be 'dir', got %v", resp[0])
	}

	dirPath, ok := resp[1].(string)
	if !ok {
		return fmt.Errorf("Expected second element in response to be a string, got %T", resp[1])
	}

	if dirPath != tmpDir {
		return fmt.Errorf("Expected second element in response to be %v, got %v", tmpDir, dirPath)
	}

	client.Close()
	return nil
}
