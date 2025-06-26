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

	randomKey := testerutils_random.RandomWord()
	entryID := "0-1"
	randomInt := strconv.Itoa(testerutils_random.RandomInt(1, 100))
	temperature := "temperature"

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"XADD", randomKey, entryID, temperature, randomInt},
			{"XREAD", "streams", randomKey, "0-0"},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion(entryID),
			resp_assertions.NewXReadResponseAssertion([]resp_assertions.StreamResponse{
				{
					Key: randomKey,
					Entries: []resp_assertions.StreamEntry{{
						Id:              entryID,
						FieldValuePairs: [][]string{{temperature, randomInt}},
					}},
				},
			}),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
