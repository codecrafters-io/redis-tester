package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubSubscribe2(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
		Logger:       logger,
	}
	clients, err := clientsSpawner.SpawnClients(2)
	if err != nil {
		return err
	}

	firstClient := clients[0]
	secondClient := clients[1]
	channelsCount := random.RandomInt(3, 6)
	channels := random.RandomWords(channelsCount)

	subscribeTestCase := test_cases.SubscriberGroupTestCase{}

	for _, c := range channels {
		subscribeTestCase.AddSubscription(firstClient, c)
		subscribeTestCase.AddSubscription(secondClient, c)
	}

	return subscribeTestCase.RunSubscribe(logger)
}
