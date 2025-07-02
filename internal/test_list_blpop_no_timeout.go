package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
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

	sendingClient := clients[0]
	firstBlockingClient := clients[1]
	secondBlockingClient := clients[2]
	blockingClients := []*resp_connection.RespConnection{firstBlockingClient, secondBlockingClient}

	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()

	// We only send commands here, not expecting responses yet
	for _, client := range blockingClients {
		if err := client.SendCommand("BLPOP", listKey, "0"); err != nil {
			return err
		}
	}

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "RPUSH",
		Args:      []string{listKey, pushValue},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}

	if err := sendCommandTestCase.Run(sendingClient, logger); err != nil {
		return err
	}

	firstBlockingClientTestCase := test_cases.ReceiveValueTestCase{
		Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue}),
	}

	if err := firstBlockingClientTestCase.Run(firstBlockingClient, logger); err != nil {
		return err
	}

	secondBlockingClientTestCase := test_cases.NoUnreadDataTestCase{}

	if err := secondBlockingClientTestCase.Run(secondBlockingClient, logger); err != nil {
		return err
	}
}
