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

	randomListKey := testerutils_random.RandomWord()
	listSize := testerutils_random.RandomInt(4, 5)
	randomList := testerutils_random.RandomWords(listSize)
	middleIndex := testerutils_random.RandomInt(1, listSize-1)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			append([]string{"RPUSH", randomListKey}, randomList...),

			// usual test cases
			{"LRANGE", randomListKey, "0", strconv.Itoa(middleIndex)},
			{"LRANGE", randomListKey, strconv.Itoa(middleIndex), strconv.Itoa(listSize - 1)},
			{"LRANGE", randomListKey, "0", strconv.Itoa(listSize - 1)},

			// start index > end index
			{"LRANGE", randomListKey, "1", "0"},

			// end index out of bounds
			{"LRANGE", randomListKey, "0", strconv.Itoa(listSize * 2)},

			// key doesn't exist
			{"LRANGE", "non_existent_key", "0", "1"},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(listSize),
			resp_assertions.NewOrderedStringArrayAssertion(randomList[0 : middleIndex+1]),
			resp_assertions.NewOrderedStringArrayAssertion(randomList[middleIndex:listSize]),
			resp_assertions.NewOrderedStringArrayAssertion(randomList[0:listSize]),
			resp_assertions.NewOrderedStringArrayAssertion([]string{}),
			resp_assertions.NewOrderedStringArrayAssertion(randomList[0:listSize]),
			resp_assertions.NewOrderedStringArrayAssertion([]string{}),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
