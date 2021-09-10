package internal

import (
	"fmt"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func antiCheatTest(harness *testerutils.StageHarness) error {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})

	logger := harness.Logger

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
