package internal

import (
	"fmt"
	testerutils "github.com/codecrafters-io/tester-utils"
	"strings"
)

func antiCheatTest(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	client := NewRedisClient()
	logger := stageHarness.Logger

	result := client.Info("server")
	if result.Err() != nil {
		return nil
	}

	str, err := result.Result()
	if err != nil {
		return nil
	}

	if !strings.HasPrefix(str, "# Server") {
		return nil
	}

	logger.Criticalf("anti-cheat (ac1) failed.")
	logger.Criticalf("Are you sure you aren't running this against the actual Redis?")
	return fmt.Errorf("anti-cheat (ac1) failed")
}
