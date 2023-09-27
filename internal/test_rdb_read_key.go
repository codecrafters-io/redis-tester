package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testRdbReadKey(stageHarness *testerutils.StageHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	randomKeyAndValue := testerutils_random.RandomWords(2)
	randomKey := randomKeyAndValue[0]
	randomValue := randomKeyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: randomKey, value: randomValue}}); err != nil {
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

	if len(resp) != 1 {
		return fmt.Errorf("Expected response to contain exactly one element, got %v", len(resp))
	}

	if resp[0] != randomKey {
		return fmt.Errorf("Expected first element of response to be %v, got %v", randomKey, resp[0])
	}

	client.Close()
	return nil
}
