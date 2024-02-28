package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/command_test"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_redis_client"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/random"

	testerutils "github.com/codecrafters-io/tester-utils"
)

// Tests 'ECHO'
func testEcho(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_redis_client.NewInstrumentedRedisClient(stageHarness, "localhost:6379", "")
	if err != nil {
		return err
	}

	randomWord := random.RandomWord()

	commandTestCase := command_test.CommandTestCase{
		Command:   "echo",
		Args:      []string{randomWord},
		Assertion: resp_assertions.NewStringValueAssertion(randomWord),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()

	return nil
}
