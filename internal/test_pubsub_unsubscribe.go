package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubUnsubscribe(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	clients, err := SpawnClients(3, "localhost:6379", stageHarness, logger)
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	publisherClient := clients[2]
	channels := random.RandomWords(3)
	messages := random.RandomStrings(2)

	pubSubTestCase := test_cases.NewPubSubTestCase()

	err = pubSubTestCase.
		// client-1 subscribes to channels[0] and channels[1]
		AddSubscription(clients[0], channels[0]).
		AddSubscription(clients[0], channels[1]).
		// client-2 subscribes to channels[1] and channels[2]
		AddSubscription(clients[1], channels[1]).
		AddSubscription(clients[1], channels[2]).
		RunSubscribeFromAll(logger)

	if err != nil {
		return err
	}

	if err := pubSubTestCase.RunPublish(channels[1], messages[0], publisherClient, logger); err != nil {
		return err
	}

	if err := pubSubTestCase.RunUnsubscribe(clients[0], channels[1], logger); err != nil {
		return err
	}

	if err := pubSubTestCase.RunPublish(channels[1], messages[1], publisherClient, logger); err != nil {
		return err
	}

	return nil
}
