package main

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func testPingPong(executable *Executable, logger *customLogger) error {
	logger.Debugf("Running program")
	if err := executable.Start(); err != nil {
		return err
	}
	defer executable.Kill()

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

	client.Close()

	return nil
}
