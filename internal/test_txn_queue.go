package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxQueue(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	var clients []*resp_connection.RespConnection

	for i := 0; i < 2; i++ {
		client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", fmt.Sprintf("client-%d", i+1))
		if err != nil {
			logFriendlyError(logger, err)
			return err
		}
		clients = append(clients, client)
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
