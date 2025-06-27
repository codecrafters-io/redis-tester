package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListRpush2(stageHarness *test_case_harness.TestCaseHarness) error {
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

	randomListKey := testerutils_random.RandomWord()
	randomElements := testerutils_random.RandomWords(3)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"RPUSH", randomListKey, randomElements[0]},
			{"RPUSH", randomListKey, randomElements[1]},
			{"RPUSH", randomListKey, randomElements[2]},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(1),
			resp_assertions.NewIntegerAssertion(2),
			resp_assertions.NewIntegerAssertion(3),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
