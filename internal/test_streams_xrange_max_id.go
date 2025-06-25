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

func testStreamsXrangeMaxID(stageHarness *test_case_harness.TestCaseHarness) error {
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

	testCase := buildMaxIDRangeTestCase(randomKey, entryCount, 2)
	return testCase.RunAll(client, logger)
}

// adds from xadd 0-1 to 0-entryCount
// adds xrange with args:  0-xrangeStartID and +
func buildMaxIDRangeTestCase(key string, entryCount int, xrangeStartID int) test_cases.MultiCommandTestCase {
	testCase := test_cases.MultiCommandTestCase{}
	addXADDCommands(&testCase, key, entryCount)

	startID := fmt.Sprintf("0-%d", xrangeStartID)

	testCase.Commands = append(testCase.Commands, []string{
		"XRANGE", key, startID, "+",
	})

	expectedEntries := createExpectedStreamEntries(xrangeStartID, entryCount)
	testCase.Assertions = append(testCase.Assertions,
		resp_assertions.NewXRangeResponseAssertion(expectedEntries),
	)

	return testCase
}
