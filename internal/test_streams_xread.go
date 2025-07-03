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

func testStreamsXread(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}

	streamKey := testerutils_random.RandomWord()
	entryID := "0-1"
	entryValue := strconv.Itoa(testerutils_random.RandomInt(1, 100))
	temperature := "temperature"

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"XADD", streamKey, entryID, temperature, entryValue},
				Assertion: resp_assertions.NewStringAssertion(entryID),
			},
			{
				Command: []string{"XREAD", "streams", streamKey, "0-0"},
				Assertion: resp_assertions.NewXReadResponseAssertion([]resp_assertions.StreamResponse{
					{
						Key: streamKey,
						Entries: []resp_assertions.StreamEntry{{
							Id:              entryID,
							FieldValuePairs: [][]string{{temperature, entryValue}},
						}},
					},
				}),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
