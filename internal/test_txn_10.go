package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxErr(stageHarness *test_case_harness.TestCaseHarness) error {
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

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"SET", "foo", "abc"},
			{"SET", "bar", "7"},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("OK"),
			resp_assertions.NewStringAssertion("OK"),
		},
	}

	if err := multiCommandTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"INCR", "foo"},
			{"INCR", "bar"},
		},
		ResultArray: []resp_value.Value{
			resp_value.NewErrorValue("ERR value is not an integer or out of range"), resp_value.NewIntegerValue(8)},
	}

	if err := transactionTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	multiCommandTestCase = test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"GET", "bar"},
			{"GET", "foo"},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("8"),
			resp_assertions.NewStringAssertion("abc"),
		},
	}

	return multiCommandTestCase.RunAll(clients[1], logger)
}
