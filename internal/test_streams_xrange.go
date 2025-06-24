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
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	randomKey := testerutils_random.RandomWord()
	entryCount := testerutils_random.RandomInt(3, 5)
	xrangeStartID := 2

	testCase := buildXrangeTestCase(randomKey, entryCount, xrangeStartID, entryCount)
	return testCase.RunAll(client, stageHarness.Logger)
}

func buildXrangeTestCase(key string, entryCount, xrangeStartID, xrangeEndID int) test_cases.MultiCommandTestCase {
	testCase := test_cases.MultiCommandTestCase{}

	addXADDCommands(&testCase, key, entryCount)

	startID := fmt.Sprintf("0-%d", xrangeStartID)
	endID := fmt.Sprintf("0-%d", xrangeEndID)
	testCase.Commands = append(testCase.Commands, []string{
		"XRANGE", key, startID, endID,
	})

	expectedEntries := createExpectedStreamEntries(xrangeStartID, xrangeEndID)
	testCase.Assertions = append(testCase.Assertions,
		resp_assertions.NewXRangeResponseAssertion(expectedEntries),
	)

	return testCase
}

var KEY_VALUE_PAIRS = [][2]string{
	{"foo", "bar"},
	{"bar", "baz"},
	{"baz", "foo"},
}

// creates XADD <key> 0-1 to 0-(count-1) <key> <value>
// key and value are taken from above in round robin fashion
func addXADDCommands(testCase *test_cases.MultiCommandTestCase, key string, entryCount int) {
	for i := 1; i <= entryCount; i++ {
		entryID := fmt.Sprintf("0-%d", i)

		pairIndex := (i - 1) % 3
		testCase.Commands = append(testCase.Commands, []string{
			"XADD", key, entryID, KEY_VALUE_PAIRS[pairIndex][0], KEY_VALUE_PAIRS[pairIndex][1],
		})

		testCase.Assertions = append(testCase.Assertions,
			resp_assertions.NewStringAssertion(entryID),
		)
	}
}

// created list of streamentry starting from
// 0-startID
// to 0-endID
// expected values are picked in round-robin fashion starting with index: (startID - 1)
func createExpectedStreamEntries(startID, endID int) []resp_assertions.StreamEntry {
	if startID < 1 || endID < startID {
		panic(fmt.Sprintf("startID > endID. startID: %d, endID: %d", startID, endID))
	}
	entriesSize := (endID - startID) + 1
	entries := make([]resp_assertions.StreamEntry, entriesSize)

	for i := range entriesSize {
		pairIndex := (startID + i - 1) % 3
		entries[i] = resp_assertions.StreamEntry{
			Id:              fmt.Sprintf("0-%d", startID+i),
			FieldValuePairs: [][]string{KEY_VALUE_PAIRS[pairIndex][:]},
		}
	}

	return entries
}
