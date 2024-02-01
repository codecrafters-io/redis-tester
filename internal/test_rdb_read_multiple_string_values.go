package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testRdbReadMultipleStringValues(stageHarness *testerutils.StageHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	keyCount := testerutils_random.RandomInt(3, 6)
	keys := testerutils_random.RandomWords(keyCount)
	values := testerutils_random.RandomWords(keyCount)

	keyValuePairs := make([]KeyValuePair, keyCount)
	for i := 0; i < keyCount; i++ {
		keyValuePairs[i] = KeyValuePair{key: keys[i], value: values[i]}
	}

	formattedKeyValuePairs := make([]string, keyCount)
	for i := 0; i < keyCount; i++ {
		formattedKeyValuePairs[i] = fmt.Sprintf("%q=%q", keys[i], values[i])
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with key-value pairs: %s", strings.Join(formattedKeyValuePairs, ", "))

	if err := RDBFileCreator.Write(keyValuePairs); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	b := NewRedisBinary(stageHarness)
	b.args = []string{
		"--dir", RDBFileCreator.Dir,
		"--dbfilename", RDBFileCreator.Filename,
	}

	if err := b.Run(); err != nil {
		return err
	}

	client := NewRedisClient("localhost:6379")

	for _, key := range keys {
		logger.Infof(fmt.Sprintf("$ redis-cli GET %s", key))
		resp, err := client.Get(key).Result()
		if err != nil {
			logFriendlyError(logger, err)
			return err
		}

		expectedValue := ""
		for _, kv := range keyValuePairs {
			if kv.key == key {
				expectedValue = kv.value
				break
			}
		}

		if resp != expectedValue {
			return fmt.Errorf("Expected response to be %v, got %v", expectedValue, resp)
		}
	}

	client.Close()
	return nil
}
