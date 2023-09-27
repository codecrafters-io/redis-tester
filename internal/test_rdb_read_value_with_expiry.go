package internal

import (
	"fmt"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

func testRdbReadValueWithExpiry(stageHarness *testerutils.StageHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	keyCount := testerutils_random.RandomInt(3, 6)
	keys := testerutils_random.RandomWords(keyCount)
	values := testerutils_random.RandomWords(keyCount)
	expiringKeyIndex := testerutils_random.RandomInt(0, keyCount-1)

	keyValuePairs := make([]KeyValuePair, keyCount)
	for i := 0; i < keyCount; i++ {
		if expiringKeyIndex == i {
			keyValuePairs[i] = KeyValuePair{
				key:      keys[i],
				value:    values[i],
				expiryTs: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
			}
		} else {
			keyValuePairs[i] = KeyValuePair{
				key:      keys[i],
				value:    values[i],
				expiryTs: time.Date(2032, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
			}
		}
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

	for keyIndex, key := range keys {
		logger.Infof(fmt.Sprintf("$ redis-cli GET %s", key))
		resp, err := client.Get(key).Result()

		if keyIndex == expiringKeyIndex {
			if err != redis.Nil {
				if err == nil {
					logger.Debugf("Hint: Read about null bulk strings in the Redis protocol docs")
					return fmt.Errorf("Expected null string, got %#v", resp)
				} else {
					logFriendlyError(logger, err)
					return err
				}
			}
		} else {
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

	}

	client.Close()
	return nil
}
