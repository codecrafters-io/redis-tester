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

func testListLlen(stageHarness *test_case_harness.TestCaseHarness) error {
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
	listSize := testerutils_random.RandomInt(4, 8)
	randomList := testerutils_random.RandomWords(listSize)
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			append([]string{"RPUSH", randomKey}, randomList...),
			{"LLEN", randomKey},
			{"LLEN", missingKey},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewIntegerAssertion(listSize),
			resp_assertions.NewIntegerAssertion(listSize),
			resp_assertions.NewIntegerAssertion(0),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
