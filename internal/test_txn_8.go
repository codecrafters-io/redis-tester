package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxSuccess(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client1")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", "foo", "6"},
			{"INCR", "foo"},
			{"INCR", "bar"},
			{"GET", "bar"},
		},
		ResultArray: []resp_value.Value{resp_value.NewSimpleStringValue("OK"), resp_value.NewIntegerValue(7), resp_value.NewIntegerValue(1), resp_value.NewBulkStringValue("1")},
	}

	return transactionTestCase.RunAll(client, logger)
}
