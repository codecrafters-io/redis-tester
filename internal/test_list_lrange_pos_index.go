package internal

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListLrangePosIdx(stageHarness *test_case_harness.TestCaseHarness) error {
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
	middleIndex := testerutils_random.RandomInt(1, listSize-1)
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   append([]string{"RPUSH", listKey}, elements...),
				Assertion: resp_assertions.NewIntegerAssertion(listSize),
			},
			// usual test cases
			{
				Command:   []string{"LRANGE", listKey, "0", strconv.Itoa(middleIndex)},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion(elements[0 : middleIndex+1]),
			},
			{
				Command:   []string{"LRANGE", listKey, strconv.Itoa(middleIndex), strconv.Itoa(listSize - 1)},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion(elements[middleIndex:listSize]),
			},
			{
				Command:   []string{"LRANGE", listKey, "0", strconv.Itoa(listSize - 1)},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion(elements[0:listSize]),
			},
			// start index > end index
			{
				Command:   []string{"LRANGE", listKey, "1", "0"},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{}),
			},
			// end index out of bounds
			{
				Command:   []string{"LRANGE", listKey, "0", strconv.Itoa(listSize * 2)},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion(elements[0:listSize]),
			},
			// key doesn't exist
			{
				Command:   []string{"LRANGE", missingKey, "0", "1"},
				Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{}),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
