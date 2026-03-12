package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingWatchInTxn(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	// Begin a transaction
	multiTestCase := test_cases.TransactionTestCase{}

	// Don't run exec yet
	if err := multiTestCase.RunWithoutExec(client, logger); err != nil {
		return err
	}

	// Watch a key
	watchTestCase := test_cases.WatchTestCase{
		Keys:                []string{random.RandomWord()},
		IsInsideTransaction: true,
	}

	return watchTestCase.Run(client, logger)
}
