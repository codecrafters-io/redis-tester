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

func testListLpop1(stageHarness *test_case_harness.TestCaseHarness) error {
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

	listKey := testerutils_random.RandomWord()
	listSize := testerutils_random.RandomInt(4, 8)
	elements := testerutils_random.RandomWords(listSize)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   append([]string{"RPUSH", listKey}, elements...),
				Assertion: resp_assertions.NewIntegerAssertion(listSize),
			},
			{
				Command:   []string{"LPOP", listKey},
				Assertion: resp_assertions.NewStringAssertion(elements[0]),
			},
			{
				Command:   []string{"LRANGE", listKey, strconv.Itoa(0), strconv.Itoa(-1)},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion(elements[1:]),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
