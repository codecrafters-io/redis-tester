package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/command_test"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_redis_client"
	"github.com/codecrafters-io/redis-tester/internal/redis_client"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	testerutils "github.com/codecrafters-io/tester-utils"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

func testPingPongOnce(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "")
	if err != nil {
		return err
	}

	logger.Debugln("Connection established, sending ping command...")

	commandTestCase := command_test.CommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewStringValueAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}

func testPingPongMultiple(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client); err != nil {
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
	client1, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}

	client2, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "client-2")
	if err != nil {
		return err
	}

	if err := runPing(logger, client2); err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}
	if err := runPing(logger, client1); err != nil {
		return err
	}
	if err := runPing(logger, client2); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 1)
	client1.Close()

	client3, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "client-3")
	if err != nil {
		return err
	}

	if err := runPing(logger, client3); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 2)
	client2.Close()
	logger.Debugf("client-%d: Success, closing connection...", 3)
	client3.Close()

	return nil
}

func runPing(logger *logger.Logger, client *redis_client.RedisClient) error {
	commandTestCase := command_test.CommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewStringValueAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}
