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

	firstClient := clients[0]
	subscribeTestCase1 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[0]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[0], 1),
	}
	if err := subscribeTestCase1.Run(firstClient, logger); err != nil {
		return err
	}

	secondClient := clients[1]
	subscribeTestCase2 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[1]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[1], 1),
	}
	if err := subscribeTestCase2.Run(secondClient, logger); err != nil {
		return err
	}

	thirdClient := clients[2]
	subscribeTestCase3 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[1]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[1], 1),
	}
	if err := subscribeTestCase3.Run(thirdClient, logger); err != nil {
		return err
	}

	publishClient := clients[3]
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
	return publishTestCase.RunAll(publishClient, logger)
}
