package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXaddPartialAutoid(stageHarness *test_case_harness.TestCaseHarness) error {
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

	randomKey := random.RandomWord()
	randomValues := random.RandomWords(3)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"XADD", randomKey, "0-*", randomValues[0], randomValues[1]},
			{"XADD", randomKey, "1-*", randomValues[0], randomValues[1]},
			{"XADD", randomKey, "1-*", randomValues[1], randomValues[2]},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("0-1"),
			resp_assertions.NewStringAssertion("1-0"),
			resp_assertions.NewStringAssertion("1-1"),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
