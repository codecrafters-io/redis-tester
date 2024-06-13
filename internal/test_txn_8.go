package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxSuccess(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, client := range clients {
		defer client.Close()
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", "foo", "6"},
			{"INCR", "foo"},
			{"INCR", "bar"},
			{"GET", "bar"},
		},
		ResultArray: []resp_value.Value{resp_value.NewSimpleStringValue("OK"), resp_value.NewIntegerValue(7), resp_value.NewIntegerValue(1), resp_value.NewBulkStringValue("1")},
	}

	if err := transactionTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{"foo"},
		Assertion: resp_assertions.NewStringAssertion("7"),
	}

	return commandTestCase.Run(clients[1], logger)
}
