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

func testListLpop2(stageHarness *test_case_harness.TestCaseHarness) error {
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
	listSize := testerutils_random.RandomInt(5, 9)
	elements := testerutils_random.RandomWords(listSize)
	toRemoveCount := testerutils_random.RandomInt(2, 5)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			append([]string{"RPUSH", randomListKey}, elements...),
			{"LPOP", randomListKey, strconv.Itoa(toRemoveCount)},
			{"LRANGE", randomListKey, strconv.Itoa(0), strconv.Itoa(-1)},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(listSize),
			resp_assertions.NewOrderedStringArrayAssertion(elements[0:toRemoveCount]),
			resp_assertions.NewOrderedStringArrayAssertion(elements[toRemoveCount:]),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
