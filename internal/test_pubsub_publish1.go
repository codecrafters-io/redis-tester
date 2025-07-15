package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
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

	err = pubSubTestCase.
		AddSubscriber(clients[0], channels[0]).
		AddSubscriber(clients[1], channels[1]).
		AddSubscriber(clients[2], channels[1]).
		SubscribeFromAll(logger)

	if err != nil {
		return err
	}

	publisherClient := clients[3]
	publishTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"PUBLISH", channels[1], "msg"},
				Assertion: resp_assertions.NewIntegerAssertion(2),
			},
			{
				Command:   []string{"PUBLISH", channels[0], "msg"},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
		},
	}
	return publishTestCase.RunAll(publisherClient, logger)
}
