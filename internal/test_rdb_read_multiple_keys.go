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

func testRdbReadMultipleKeys(stageHarness *test_case_harness.TestCaseHarness) error {
	RDBFileCreator, err := NewRDBFileCreator()
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	keyCount := testerutils_random.RandomInt(3, 6)
	keys := testerutils_random.RandomWords(keyCount)
	values := testerutils_random.RandomWords(keyCount)

	keyValuePairs := make([]KeyValuePair, keyCount)
	for i := 0; i < keyCount; i++ {
		keyValuePairs[i] = KeyValuePair{key: keys[i], value: values[i]}
	}

	if err := RDBFileCreator.Write(keyValuePairs); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with %d keys: %v", keyCount, FormatKeys(keys))
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
		Assertion:                 resp_assertions.NewUnorderedStringArrayAssertion(keys),
		ShouldSkipUnreadDataCheck: false,
	}

	return commandTestCase.Run(client, logger)
}
