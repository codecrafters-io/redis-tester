package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testRdbReadStringValue(stageHarness *testerutils.StageHarness) error {
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

	logger.Infof(fmt.Sprintf("$ redis-cli GET %s", randomKey))
	resp, err := client.Get(randomKey).Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != randomValue {
		return fmt.Errorf("Expected response to be %v, got %v", randomValue, resp)
	}

	client.Close()
	return nil
}
