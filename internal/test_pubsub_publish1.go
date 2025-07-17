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

	subscriberGroupTestCase := test_cases.SubscriberGroupTestCase{}
	subscriberGroupTestCase.
		AddSubscription(clients[0], channels[0]).
		AddSubscription(clients[1], channels[1]).
		AddSubscription(clients[2], channels[1])

	err = subscriberGroupTestCase.RunSubscribe(logger)
	if err != nil {
		return err
	}

	messages := random.RandomWords(2)
	publisherClient := clients[3]

	publishTestCase1 := test_cases.PublishTestCase{
		Channel:                 channels[0],
		Message:                 messages[0],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[0]),
	}
	if err := publishTestCase1.Run(publisherClient, logger); err != nil {
		return err
	}

	publishTestCase2 := test_cases.PublishTestCase{
		Channel:                 channels[1],
		Message:                 messages[1],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[1]),
	}
	return publishTestCase2.Run(publisherClient, logger)
}
