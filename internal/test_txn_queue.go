package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxQueue(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := spawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, client := range clients {
		defer client.Close()
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", "foo", "41"},
			{"INCR", "foo"},
		},
		ResultArray: []resp_value.Value{},
	}

	if err := transactionTestCase.RunMulti(clients[0], logger); err != nil {
		return err
	}

	if err := transactionTestCase.RunQueueAll(clients[0], logger); err != nil {
		return err
	}

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{"foo"},
		Assertion: resp_assertions.NewNilAssertion(),
	}

	return commandTestCase.Run(clients[1], logger)
}
