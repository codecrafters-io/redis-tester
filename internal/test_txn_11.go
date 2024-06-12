package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxMulti(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	var clients []*resp_connection.RespConnection

	for i := 0; i < 3; i++ {
		client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client1")
		if err != nil {
			logFriendlyError(logger, err)
			return err
		}
		clients = append(clients, client)
		defer client.Close()
	}

	for i, client := range clients {
		commandTestCase := test_cases.SendCommandTestCase{
			Command:   "SET",
			Args:      []string{"bar", "7"},
			Assertion: resp_assertions.NewStringAssertion("OK"),
		}

		if err := commandTestCase.Run(client, logger); err != nil {
			return err
		}

		commandTestCase = test_cases.SendCommandTestCase{
			Command:   "INCR",
			Args:      []string{"foo"},
			Assertion: resp_assertions.NewIntegerAssertion(i + 1),
		}

		if err := commandTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	for _, client := range clients {
		transactionTestCase := test_cases.TransactionTestCase{
			CommandQueue: [][]string{
				{"INCR", "foo"},
				{"INCR", "bar"},
			},
			ResultArray: []resp_value.Value{},
		}
		if err := transactionTestCase.RunAll(client, logger); err != nil {
			return err
		}
	}

	for i, client := range clients {
		transactionTestCase := test_cases.TransactionTestCase{
			CommandQueue: [][]string{},
			ResultArray:  []resp_value.Value{resp_value.NewIntegerValue(4 + i), resp_value.NewIntegerValue(8 + i)},
		}
		if err := transactionTestCase.RunExec(client, logger); err != nil {
			return err
		}
	}

	return nil
}
