package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
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

	value, err := client.ReadValue()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if err = resp_assertions.NewStringValueAssertion("PONG").Run(value); err != nil {
		return err
	}

	if client.UnreadBuffer.Len() > 0 {
		return fmt.Errorf("Found extra data: %q", string(client.LastValueBytes)+client.UnreadBuffer.String())
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
