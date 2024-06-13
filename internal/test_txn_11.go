package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxMultiTx(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := SpawnClients(3, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, client := range clients {
		defer client.Close()
	}

	for i, client := range clients {
		multiCommandTestCase := test_cases.MultiCommandTestCase{
			Commands: [][]string{
				{"SET", "bar", "7"},
				{"INCR", "foo"},
			},
			Assertions: []resp_assertions.RESPAssertion{
				resp_assertions.NewStringAssertion("OK"),
				resp_assertions.NewIntegerAssertion(i + 1),
			},
		}

		if err := multiCommandTestCase.RunAll(client, logger); err != nil {
			return err
		}
	}

	for i, client := range clients {
		transactionTestCase := test_cases.TransactionTestCase{
			CommandQueue: [][]string{
				{"INCR", "foo"},
				{"INCR", "bar"},
			},
			ResultArray: []resp_value.Value{resp_value.NewIntegerValue(4 + i), resp_value.NewIntegerValue(8 + i)},
		}
		if err := transactionTestCase.RunMulti(client, logger); err != nil {
			return err
		}

		if err := transactionTestCase.RunQueueAll(client, logger); err != nil {
			return err
		}
	}

	for i, client := range clients {
		transactionTestCase := test_cases.TransactionTestCase{
			CommandQueue: [][]string{
				{"INCR", "foo"},
				{"INCR", "bar"},
			},
			ResultArray: []resp_value.Value{resp_value.NewIntegerValue(4 + i), resp_value.NewIntegerValue(8 + i)},
		}
		if err := transactionTestCase.RunExec(client, logger); err != nil {
			return err
		}
	}

	return nil
}
