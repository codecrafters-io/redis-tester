package internal

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbConfig(stageHarness *test_case_harness.TestCaseHarness) error {
	tmpDir, err := MkdirTemp("rdb")
	if err != nil {
		return err
	}

	// On MacOS, the tmpDir is a symlink to a directory in /var/folders/...
	b := redis_executable.NewRedisExecutable(stageHarness)
	stageHarness.RegisterTeardownFunc(func() { os.RemoveAll(tmpDir) })
	if err := b.Run("--dir", tmpDir,
		"--dbfilename", fmt.Sprintf("%s.rdb", testerutils_random.RandomWord())); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}
	defer client.Close()

	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "CONFIG",
		Args:                      []string{"GET", "dir"},
		Assertion:                 resp_assertions.NewOrderedStringArrayAssertion([]string{"dir", tmpDir}),
		ShouldSkipUnreadDataCheck: false,
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}
