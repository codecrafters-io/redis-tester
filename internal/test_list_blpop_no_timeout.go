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

	blPopAssertion := resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue})

	blockingTestCase := test_cases.BlockingCommandTestCase{
		BlockingClientsTestCases: []test_cases.ClientTestCase{
			{
				Client: clients[0],
				SendCommandTestCase: &test_cases.SendCommandTestCase{
					Command:   "BLPOP",
					Args:      []string{listKey, "0"},
					Assertion: blPopAssertion,
				},
				ExpectResult: true,
			},
			{
				Client: clients[1],
				SendCommandTestCase: &test_cases.SendCommandTestCase{
					Command:   "BLPOP",
					Args:      []string{listKey, "0"},
					Assertion: blPopAssertion,
				},
				ExpectResult: false,
			},
		},
		UnblockingClientTestCase: &test_cases.ClientTestCase{
			Client: clients[2],
			SendCommandTestCase: &test_cases.SendCommandTestCase{
				Command:   "RPUSH",
				Args:      []string{listKey, pushValue},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			ExpectResult: true,
		},
	}

	return blockingTestCase.Run(logger)
}
