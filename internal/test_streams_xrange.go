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

	streamKey := testerutils_random.RandomWord()

	entryCount := testerutils_random.RandomInt(3, 5)
	var entryIDs []string
	for i := range entryCount {
		entryIDs = append(entryIDs, fmt.Sprintf("0-%d", i+1))
	}

	entryKeysAndValues := make([][]string, entryCount)
	for i := range entryCount {
		entryKeysAndValues[i] = testerutils_random.RandomWords(2)
	}

	commandWithAssertions := []test_cases.CommandWithAssertion{}
	for i := range entryCount {
		commandWithAssertions = append(commandWithAssertions, test_cases.CommandWithAssertion{
			Command:   []string{"XADD", streamKey, entryIDs[i], entryKeysAndValues[i][0], entryKeysAndValues[i][1]},
			Assertion: resp_assertions.NewStringAssertion(entryIDs[i]),
		})
	}

	xaddTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: commandWithAssertions,
	}

	if err := xaddTestCase.RunAll(client, logger); err != nil {
		return err
	}

	startKey := 2
	expectedStreamEntries := []resp_assertions.StreamEntry{}
	for i := startKey - 1; i < entryCount; i++ {
		expectedStreamEntries = append(expectedStreamEntries, resp_assertions.StreamEntry{
			Id:              entryIDs[i],
			FieldValuePairs: [][]string{entryKeysAndValues[i]},
		})
	}

	xrangeTestCase := test_cases.SendCommandTestCase{
		Command:                   "XRANGE",
		Args:                      []string{streamKey, fmt.Sprintf("0-%d", startKey), fmt.Sprintf("0-%d", entryCount)},
		Assertion:                 resp_assertions.NewXRangeResponseAssertion(expectedStreamEntries),
		ShouldSkipUnreadDataCheck: false,
	}

	return xrangeTestCase.Run(client, logger)
}
