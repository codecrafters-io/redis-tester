package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
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
	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	firstClient := clients[0]
	secondClient := clients[1]
	channelsCount := random.RandomInt(3, 6)
	channels := random.RandomWords(channelsCount)

	subscribeTestCase := test_cases.MultiCommandTestCase{}

	for i, c := range channels {
		subscribeTestCase.CommandWithAssertions = append(subscribeTestCase.CommandWithAssertions,
			test_cases.CommandWithAssertion{
				Command:   []string{"SUBSCRIBE", c},
				Assertion: resp_assertions.NewSubscribeResponseAssertion(c, i+1),
			},
		)
	}

	if err := subscribeTestCase.RunAll(firstClient, logger); err != nil {
		return err
	}

	/* Run using the another client to check if subscribe counts are maintained on per-client basis */
	return subscribeTestCase.RunAll(secondClient, logger)
}
