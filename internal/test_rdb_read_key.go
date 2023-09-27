package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testRdbReadKey(stageHarness *testerutils.StageHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	if err := RDBFileCreator.Write([]KeyValuePair{{key: "hello", value: "world"}}); err != nil {
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

	logger.Infof("$ redis-cli GET hello")
	resp, err := client.Get("hello").Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "world" {
		return fmt.Errorf("Expected response to be 'world', got %v", resp)
	}

	client.Close()
	return nil
}
