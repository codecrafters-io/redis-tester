package internal

import (
	"fmt"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

func testPingPongOnce(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	logger.Debugf("Sending ping command...")
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}

	if pong != "PONG" {
		return fmt.Errorf("Expected PONG, got %s", pong)
	}

	logger.Debugf("Success, closing connection...")
	client.Close()

	return nil
}

func testPingPongMultiple(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client, 1); err != nil {
			return err
		}
	}

	logger.Debugf("Success, closing connection...")
	client.Close()

	return nil
}

func testPingPongConcurrent(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client1 := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	if err := runPing(logger, client1, 1); err != nil {
		return err
	}

	client2 := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	if err := runPing(logger, client2, 2); err != nil {
		return err
	}

	if err := runPing(logger, client1, 1); err != nil {
		return err
	}
	if err := runPing(logger, client1, 1); err != nil {
		return err
	}
	if err := runPing(logger, client2, 2); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 1)
	client1.Close()

	client3 := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})
	if err := runPing(logger, client3, 3); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 2)
	client2.Close()
	logger.Debugf("client-%d: Success, closing connection...", 3)
	client3.Close()

	return nil
}

func runPing(logger *testerutils.Logger, client *redis.Client, clientNum int) error {
	logger.Debugf("client-%d: Sending ping command...", clientNum)
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}

	logger.Debugf("client-%d: Received response.", clientNum)

	if pong != "PONG" {
		return fmt.Errorf("client-%d: Expected PONG, got %s", clientNum, pong)
	}

	return nil
}
