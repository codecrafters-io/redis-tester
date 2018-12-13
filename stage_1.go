package main

import "fmt"
import "github.com/go-redis/redis"
import "time"

func getStageOne() Stage {
	return Stage{
		name:        "stage 1",
		description: "stage 1 desc",
		runFunc:     runStage1,
	}
}

func runStage1() error {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}

	if pong != "PONG" {
		return fmt.Errorf("Expected PONG, got %s", pong)
	}

	return nil
}
