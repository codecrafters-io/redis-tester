package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListBlpopNoTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := SpawnClients(3, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()

	blPopResponseAssertion := resp_assertions.NewOrderedBulkStringArrayAssertion([]string{listKey, pushValue})

	blockingClientGroupTestCase := test_cases.BlockingClientGroupTestCase{
		CommandToSend:                 []string{"BLPOP", listKey, "0"},
		AssertionForReceivedResponse:  blPopResponseAssertion,
		ResponseExpectingClientsCount: 1,
		Clients:                       clients[0:2],
	}

	// We only send commands here, not expecting responses yet
	if err := blockingClientGroupTestCase.SendBlockingCommands(); err != nil {
		return err
	}

	rpushTestCase := test_cases.SendCommandTestCase{
		Command:   "RPUSH",
		Args:      []string{listKey, pushValue},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}

	sendingClient := clients[2]
	if err := rpushTestCase.Run(sendingClient, logger); err != nil {
		return err
	}

	return blockingClientGroupTestCase.AssertResponses(logger)
}
