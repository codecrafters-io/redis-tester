package internal

import (
	"fmt"
	"sort"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testRdbReadMultipleKeys(stageHarness *testerutils.StageHarness) error {
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

	logger := stageHarness.Logger
	client := NewRedisClient()

	logger.Infof("$ redis-cli KEYS *")
	resp, err := client.Keys("*").Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if len(resp) != len(keys) {
		return fmt.Errorf("Expected response to contain exactly %v elements, got %v", len(keys), len(resp))
	}

	expectedKeysSorted := make([]string, len(keys))
	copy(expectedKeysSorted, keys)
	sort.Strings(expectedKeysSorted)

	actualKeysSorted := make([]string, len(resp))
	copy(actualKeysSorted, resp)
	sort.Strings(actualKeysSorted)

	if fmt.Sprintf("%v", actualKeysSorted) != fmt.Sprintf("%v", expectedKeysSorted) {
		return fmt.Errorf("Expected response to be %v, got %v (sorted alphabetically for comparison)", expectedKeysSorted, actualKeysSorted)
	}

	client.Close()
	return nil
}
