package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxIncr2(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client, err := clientsSpawner.SpawnClientWithPrefix("client")
	if err != nil {
		return err
	}

	key := random.RandomWord()

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"INCR", key},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			{
				Command:   []string{"INCR", key},
				Assertion: resp_assertions.NewIntegerAssertion(2),
			},
			{
				Command:   []string{"GET", key},
				Assertion: resp_assertions.NewBulkStringAssertion("2"),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
