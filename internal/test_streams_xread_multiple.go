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

func testStreamsXreadMultiple(stageHarness *test_case_harness.TestCaseHarness) error {
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

	randomKeys := testerutils_random.RandomWords(2)
	entryIDs := []string{"0-1", "0-2"}

	var randomInts []string
	for i := range testerutils_random.RandomInts(1, 100, 2) {
		randomInts = append(randomInts, fmt.Sprintf("%d", i))
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"XADD", randomKeys[0], entryIDs[0], "temperature", randomInts[0]},
				Assertion: resp_assertions.NewStringAssertion(entryIDs[0]),
			},
			{
				Command:   []string{"XADD", randomKeys[1], entryIDs[1], "humidity", randomInts[1]},
				Assertion: resp_assertions.NewStringAssertion(entryIDs[1]),
			},
			{
				Command: []string{"XREAD", "streams", randomKeys[0], randomKeys[1], "0-0", "0-1"},
				Assertion: resp_assertions.NewXReadResponseAssertion([]resp_assertions.StreamResponse{
					{
						Key: randomKeys[0],
						Entries: []resp_assertions.StreamEntry{{
							Id:              entryIDs[0],
							FieldValuePairs: [][]string{{"temperature", randomInts[0]}},
						}},
					},
					{
						Key: randomKeys[1],
						Entries: []resp_assertions.StreamEntry{{
							Id:              entryIDs[1],
							FieldValuePairs: [][]string{{"humidity", randomInts[1]}},
						}},
					},
				}),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
