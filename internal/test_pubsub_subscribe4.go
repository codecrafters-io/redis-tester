package internal

import (
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
	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	for _, c := range clients {
		defer c.Close()

	}

	channel := random.RandomWord()

	subscribeTestCase := test_cases.SubscriberGroupTestCase{}
	subscribeTestCase.AddSubscription(clients[0], channel)
	if err := subscribeTestCase.RunSubscribe(logger); err != nil {
		return err
	}

	/* Test against Ping */
	pingTestCase1 := test_cases.SendCommandTestCase{
		Command:   "PING",
		Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{"pong", ""}),
	}
	if err := pingTestCase1.Run(clients[0], logger); err != nil {
		return err
	}

	/* Test against ping from a separate (unsubscribed) client */
	pingTestCase2 := test_cases.SendCommandTestCase{
		Command:   "PING",
		Assertion: resp_assertions.NewSimpleStringAssertion("PONG"),
	}
	return pingTestCase2.Run(clients[1], logger)
}
