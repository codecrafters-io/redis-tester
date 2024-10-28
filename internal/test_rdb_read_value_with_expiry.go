package internal

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbReadValueWithExpiry(stageHarness *test_case_harness.TestCaseHarness) error {
	RDBFileCreator, err := NewRDBFileCreator()
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	keyCount := testerutils_random.RandomInt(3, 6)
	keys := testerutils_random.RandomWords(keyCount)
	values := testerutils_random.RandomWords(keyCount)
	expiringKeyIndex := testerutils_random.RandomInt(0, keyCount-1)

	keyValueMap := make(map[string]string)
	keyValuePairs := make([]KeyValuePair, keyCount)
	for i := 0; i < keyCount; i++ {
		if expiringKeyIndex == i {
			keyValuePairs[i] = KeyValuePair{
				key:      keys[i],
				value:    values[i],
				expiryTS: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
			}
		} else {
			keyValuePairs[i] = KeyValuePair{
				key:      keys[i],
				value:    values[i],
				expiryTS: time.Date(2032, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
			}
		}
		key, value := keys[i], values[i]
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

	for keyIndex, key := range keys {
		if keyIndex == expiringKeyIndex {
			commandTestCase := test_cases.SendCommandTestCase{
				Command:                   "GET",
				Args:                      []string{key},
				Assertion:                 resp_assertions.NewNilAssertion(),
				ShouldSkipUnreadDataCheck: false,
			}
			if err := commandTestCase.Run(client, logger); err != nil {
				return err
			}
		} else {
			commandTestCase := test_cases.SendCommandTestCase{
				Command:                   "GET",
				Args:                      []string{key},
				Assertion:                 resp_assertions.NewStringAssertion(keyValueMap[key]),
				ShouldSkipUnreadDataCheck: false,
			}
			if err := commandTestCase.Run(client, logger); err != nil {
				return err
			}
		}
	}

	return nil
}
