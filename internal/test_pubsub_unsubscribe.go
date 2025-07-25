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
	messages := random.RandomWords(2)

	subscriberGroupTestCase := test_cases.SubscriberGroupTestCase{}

	// client-1 subscribes to channels[0] and channels[1]
	// client-2 subscribes to channels[1] and channels[2]
	subscriberGroupTestCase.
		AddSubscription(clients[0], channels[0]).
		AddSubscription(clients[0], channels[1]).
		AddSubscription(clients[1], channels[1]).
		AddSubscription(clients[1], channels[2])

	if err := subscriberGroupTestCase.RunSubscribe(logger); err != nil {
		return err
	}

	// publish and assert message
	publishTestCase1 := test_cases.PublishTestCase{
		Channel:                 channels[1],
		Message:                 messages[0],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[1]),
	}
	if err := publishTestCase1.Run(publisherClient, logger); err != nil {
		return err
	}
	if err := subscriberGroupTestCase.RunAssertionForPublishedMessage(channels[1], messages[0], logger); err != nil {
		return err
	}

	// unsubscribe
	subscriberGroupTestCase.RemoveSubscription(clients[0], channels[1])
	unsubscribeTestCase := test_cases.UnsubscribeTestCase{
		Channel:                                 channels[1],
		ExpectedSubscriberCountAfterUnsubscribe: subscriberGroupTestCase.GetSubscriberCount(channels[1]),
	}
	if err := unsubscribeTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	// publish and assert message
	publishTestCase2 := test_cases.PublishTestCase{
		Channel:                 channels[1],
		Message:                 messages[1],
		ExpectedSubscriberCount: subscriberGroupTestCase.GetSubscriberCount(channels[1]),
	}
	if err := publishTestCase2.Run(publisherClient, logger); err != nil {
		return err
	}
	if err := subscriberGroupTestCase.RunAssertionForPublishedMessage(channels[1], messages[1], logger); err != nil {
		return err
	}

	return nil
}
