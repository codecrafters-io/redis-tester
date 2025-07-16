package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubPublish1(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	clients, err := SpawnClients(4, "localhost:6379", stageHarness, logger)
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	channels := random.RandomWords(2)

	/*
		client-1 subscribes to channels[0]
		client-2 and client-3 subscribe to channels[1]
	*/

	pubSubTestCase := test_cases.NewPubSubTestCase()
	pubSubTestCase.
		AddSubscription(clients[0], channels[0]).
		AddSubscription(clients[1], channels[1]).
		AddSubscription(clients[2], channels[1])

	err = pubSubTestCase.RunSubscribeFromAll(logger)
	if err != nil {
		return err
	}

	messages := random.RandomWords(2)
	publisherClient := clients[3]

	err = pubSubTestCase.RunPublishWithoutMessageAssertion(channels[1], messages[0], publisherClient, logger)
	if err != nil {
		return err
	}

	return pubSubTestCase.RunPublishWithoutMessageAssertion(channels[0], messages[1], publisherClient, logger)
}
