package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXaddValidateID(stageHarness *test_case_harness.TestCaseHarness) error {
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

	streamKey := random.RandomWord()
	entryKeyAndValues := random.RandomWords(10)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"XADD", streamKey, "1-1", entryKeyAndValues[0], entryKeyAndValues[1]},
				Assertion: resp_assertions.NewStringAssertion("1-1"),
			},
			{
				Command:   []string{"XADD", streamKey, "1-2", entryKeyAndValues[2], entryKeyAndValues[3]},
				Assertion: resp_assertions.NewStringAssertion("1-2"),
			},
			{
				Command:   []string{"XADD", streamKey, "1-2", entryKeyAndValues[4], entryKeyAndValues[5]},
				Assertion: resp_assertions.NewErrorAssertion("ERR The ID specified in XADD is equal or smaller than the target stream top item"),
			},
			{
				Command:   []string{"XADD", streamKey, "0-3", entryKeyAndValues[6], entryKeyAndValues[7]},
				Assertion: resp_assertions.NewErrorAssertion("ERR The ID specified in XADD is equal or smaller than the target stream top item"),
			},
			{
				Command:   []string{"XADD", streamKey, "0-0", entryKeyAndValues[8], entryKeyAndValues[9]},
				Assertion: resp_assertions.NewErrorAssertion("ERR The ID specified in XADD must be greater than 0-0"),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
