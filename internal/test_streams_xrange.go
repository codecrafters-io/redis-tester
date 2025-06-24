package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXrange(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}

	randomKey := testerutils_random.RandomWord()
	entryCount := testerutils_random.RandomInt(3, 5)
	const startIDNum = 2

	testCase := buildXrangeTestCase(randomKey, entryCount, startIDNum, entryCount)
	return testCase.RunAll(client, stageHarness.Logger)
}

func buildXrangeTestCase(key string, entryCount, startIDNum, endIDNum int) test_cases.MultiCommandTestCase {
	testCase := test_cases.MultiCommandTestCase{}

	addXADDCommands(&testCase, key, entryCount)

	startID := fmt.Sprintf("0-%d", startIDNum)
	endID := fmt.Sprintf("0-%d", endIDNum)
	testCase.Commands = append(testCase.Commands, []string{
		"XRANGE", key, startID, endID,
	})

	expectedEntries := createExpectedStreamEntries(startIDNum, endIDNum)
	testCase.Assertions = append(testCase.Assertions,
		resp_assertions.NewXRangeResponseAssertion(expectedEntries),
	)

	return testCase
}

func addXADDCommands(testCase *test_cases.MultiCommandTestCase, key string, entryCount int) {
	for i := 1; i <= entryCount; i++ {
		entryID := fmt.Sprintf("0-%d", i)

		testCase.Commands = append(testCase.Commands, []string{
			"XADD", key, entryID, "foo", "bar",
		})

		testCase.Assertions = append(testCase.Assertions,
			resp_assertions.NewStringAssertion(entryID),
		)
	}
}

func createExpectedStreamEntries(startIDNum, endIDNum int) []resp_assertions.StreamEntry {
	entriesSize := (endIDNum - startIDNum) + 1
	entries := make([]resp_assertions.StreamEntry, entriesSize)

	for i := range entriesSize {
		entries[i] = resp_assertions.StreamEntry{
			Id:              fmt.Sprintf("0-%d", startIDNum+i),
			FieldValuePairs: [][]string{{"foo", "bar"}},
		}
	}

	return entries
}
