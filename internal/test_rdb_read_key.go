package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbReadKey(stageHarness *test_case_harness.TestCaseHarness) error {
	RDBFileCreator, err := NewRDBFileCreator()
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	keyAndValue := testerutils_random.RandomWords(2)
	key, value := keyAndValue[0], keyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: key, value: value}}); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with a single key: [%q]", key)
	if err := RDBFileCreator.PrintContentHexdump(logger); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	b := redis_executable.NewRedisExecutable(stageHarness)
	stageHarness.RegisterTeardownFunc(func() { RDBFileCreator.Cleanup() })
	if err := b.Run("--dir", RDBFileCreator.Dir,
		"--dbfilename", RDBFileCreator.Filename); err != nil {
		return err
	}

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}
	defer client.Close()

	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "KEYS",
		Args:                      []string{"*"},
		Assertion:                 resp_assertions.NewCommandAssertion(key),
		ShouldSkipUnreadDataCheck: false,
	}

	return commandTestCase.Run(client, logger)
}
