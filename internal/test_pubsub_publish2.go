package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubPublish2(stageHarness *test_case_harness.TestCaseHarness) error {
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

	/*
		client-1 and client-2 subscribe to channel[0]
		client-3 subscribes to channel[1]
		client-4 publishes separate messages for channel[0] and channel[1]
	*/
	publisherClient := clients[3]
	channels := random.RandomWords(2)
	messages := random.RandomWords(2)

	subscriberGroupTestCase := test_cases.SubscriberGroupTestCase{}
	subscriberGroupTestCase.
		AddSubscription(clients[0], channels[0]).
		AddSubscription(clients[1], channels[0]).
		AddSubscription(clients[2], channels[1])

	if err := subscriberGroupTestCase.RunSubscribe(logger); err != nil {
		return err
	}

	/* Publish message and assert received messages */
	publishTestCase1 := test_cases.PublishTestCase{
		Channel:                 channels[0],
		Message:                 messages[0],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[0]),
	}
	if err := publishTestCase1.Run(publisherClient, logger); err != nil {
		return err
	}
	if err := subscriberGroupTestCase.RunAssertionForPublishedMessage(channels[0], messages[0], logger); err != nil {
		return err
	}

	/* Publish message on another channel and assert for received messages */
	publishTestCase2 := test_cases.PublishTestCase{
		Channel:                 channels[1],
		Message:                 messages[1],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[1]),
	}
	if err := publishTestCase2.Run(publisherClient, logger); err != nil {
		return err
	}
	return subscriberGroupTestCase.RunAssertionForPublishedMessage(channels[1], messages[1], logger)
}
