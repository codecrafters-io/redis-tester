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

func testRdbReadMultipleStringValues(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keyValueMap := make(map[string]string)
	for i := 0; i < keyCount; i++ {
		key := keys[i]
		value := values[i]
		keyValueMap[key] = value
	}

	if err := RDBFileCreator.Write(keyValuePairs); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with %d key-value pairs: %s", len(keys), FormatKeyValuePairs(keys, values))
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

	for i, key := range keys {
		expectedValue := keyValueMap[key]

		commandTestCase := test_cases.SendCommandTestCase{
			Command:                   "GET",
			Args:                      []string{key},
			Assertion:                 resp_assertions.NewStringAssertion(expectedValue),
			ShouldSkipUnreadDataCheck: i < len(keys)-1,
		}

		if err := commandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
