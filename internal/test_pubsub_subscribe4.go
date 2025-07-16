package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubSubscribe4(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	channel := random.RandomWord()

	subscribeTestCase := test_cases.NewPubSubTestCase()
	subscribeTestCase.AddSubscription(client, channel)
	if err := subscribeTestCase.RunSubscribeFromAll(logger); err != nil {
		return err
	}

	/* Test against Ping */
	pingTestCase := test_cases.SendCommandTestCase{
		Command:   "PING",
		Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{"pong", ""}),
	}
	return pingTestCase.Run(client, logger)
}
