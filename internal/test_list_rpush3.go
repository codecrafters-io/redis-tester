package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListRpush3(stageHarness *test_case_harness.TestCaseHarness) error {
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

	listKey := testerutils_random.RandomWord()

	elementsForFirstCmd := testerutils_random.RandomWords(2)
	elementsForSecondCmd := testerutils_random.RandomWords(3)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   append([]string{"RPUSH", listKey}, elementsForFirstCmd...),
				Assertion: resp_assertions.NewIntegerAssertion(2),
			},
			{
				Command:   append([]string{"RPUSH", listKey}, elementsForSecondCmd...),
				Assertion: resp_assertions.NewIntegerAssertion(5),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
