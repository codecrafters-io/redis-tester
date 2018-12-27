package main

import "github.com/go-redis/redis"
import "fmt"
import "time"

func antiCheatRunner() StageRunner {
	return StageRunner{
		isDebug: false,
		stages: []Stage{
			Stage{
				name:    "AC1",
				logger:  getQuietLogger("[anticheat] "),
				runFunc: testCommand,
			},
		},
	}
}

func testCommand(logger *customLogger) error {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	result := client.Info()
	if result.Err() == nil {
		logger.Criticalf("anti-cheat (ac1) failed. ")
		logger.Criticalf(
			"Are you sure you aren't running this " +
				"against the actual Redis?")
		return fmt.Errorf("anti-cheat (ac1) failed")
	}

	return nil
}
