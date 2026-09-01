package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReadBitsAsString(stageHarness *test_case_harness.TestCaseHarness) error {
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

	key := testerutils_random.RandomWord()

	// 'B' is 0b01000010, so SETBIT at offsets 1 and 6.
	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SETBIT", key, "1", "1"},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"SETBIT", key, "6", "1"},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"GET", key},
				Assertion: resp_assertions.NewBulkStringAssertion("B"),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
