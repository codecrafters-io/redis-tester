package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListLpush(stageHarness *test_case_harness.TestCaseHarness) error {
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
	randomValues := testerutils_random.RandomWords(3)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"LPUSH", randomListKey, randomValues[2]},
			{"LPUSH", randomListKey, randomValues[1], randomValues[0]},
			{"LRANGE", randomListKey, strconv.Itoa(0), strconv.Itoa(-1)},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(1),
			resp_assertions.NewIntegerAssertion(3),
			resp_assertions.NewOrderedStringArrayAssertion([]string{randomValues[0], randomValues[1], randomValues[2]}),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
