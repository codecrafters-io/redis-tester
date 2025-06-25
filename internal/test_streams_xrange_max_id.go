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

	randomStreamKey := testerutils_random.RandomWord()

	entryCount := testerutils_random.RandomInt(3, 5)
	var entryIDs []string
	for i := range entryCount {
		entryIDs = append(entryIDs, fmt.Sprintf("0-%d", i+1))
	}

	randomPairs := make([][]string, entryCount)
	for i := range entryCount {
		randomPairs[i] = testerutils_random.RandomWords(2)
	}

	commands := [][]string{}
	assertions := []resp_assertions.RESPAssertion{}
	for i := range entryCount {
		commands = append(commands, []string{"XADD", randomStreamKey, entryIDs[i], randomPairs[i][0], randomPairs[i][1]})
		assertions = append(assertions, resp_assertions.NewStringAssertion(entryIDs[i]))
	}

	xaddTestCase := test_cases.MultiCommandTestCase{
		Commands:   commands,
		Assertions: assertions,
	}

	if err := xaddTestCase.RunAll(client, logger); err != nil {
		return err
	}

	// start at either 0-1 or 0-2
	startkey := testerutils_random.RandomInt(1, 3)

	expectedStreamEntries := []resp_assertions.StreamEntry{}
	for i := startkey; i <= entryCount; i++ {
		expectedStreamEntries = append(expectedStreamEntries, resp_assertions.StreamEntry{
			Id:              entryIDs[i-1],
			FieldValuePairs: [][]string{randomPairs[i-1]},
		})
	}

	xrangeTestCase := test_cases.SendCommandTestCase{
		Command:                   "XRANGE",
		Args:                      []string{randomStreamKey, fmt.Sprintf("0-%d", startkey), "+"},
		Assertion:                 resp_assertions.NewXRangeResponseAssertion(expectedStreamEntries),
		ShouldSkipUnreadDataCheck: false,
	}

	return xrangeTestCase.Run(client, logger)
}
