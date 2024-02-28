package internal

import (
	"fmt"

	resp "github.com/codecrafters-io/redis-tester/internal/resp"
	testerutils "github.com/codecrafters-io/tester-utils"
	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/go-redis/redis"
)

func testPingPongOnce(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := NewInstrumentedRedisClient(stageHarness, "localhost:6379")
	if err != nil {
		return err
	}

	logger.Debugln("Connection established, sending ping command...")
	if err := client.SendCommand("ping"); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Debugln("Reading response...")

	message, err := client.ReadMessage()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if message.Type != resp.SIMPLE_STRING {
		return fmt.Errorf("Expected simple string, got %#v", message.Type)
	}

	if message.String() != "PONG" {
		return fmt.Errorf("Expected \"PONG\", got %#v", message.String())
	}

	return nil
}

func testPingPongMultiple(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

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
	client1 := NewRedisClient("localhost:6379")

	if err := runPing(logger, client1, 1); err != nil {
		return err
	}

	client2 := NewRedisClient("localhost:6379")
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

	client3 := NewRedisClient("localhost:6379")
	if err := runPing(logger, client3, 3); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 2)
	client2.Close()
	logger.Debugf("client-%d: Success, closing connection...", 3)
	client3.Close()

	return nil
}

func runPing(logger *logger.Logger, client *redis.Client, clientNum int) error {
	logger.Infof("client-%d: Sending ping command...", clientNum)
	pong, err := client.Ping().Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if pong != "PONG" {
		logger.Debugf("client-%d: Received response.", clientNum)
		return fmt.Errorf("client-%d: Expected \"PONG\", got %#v", clientNum, pong)
	}

	logger.Successf("client-%d: Received PONG.", clientNum)

	return nil
}
