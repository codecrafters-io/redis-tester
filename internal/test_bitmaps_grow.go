package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsGrow(stageHarness *test_case_harness.TestCaseHarness) error {
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
	firstByteOffset := strconv.Itoa(testerutils_random.RandomInt(0, 8))
	thirdByteOffset := strconv.Itoa(testerutils_random.RandomInt(16, 24))

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SETBIT", key, firstByteOffset, "1"},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"STRLEN", key},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			{
				Command:   []string{"SETBIT", key, thirdByteOffset, "1"},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"GETBIT", key, thirdByteOffset},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			{
				Command:   []string{"STRLEN", key},
				Assertion: resp_assertions.NewIntegerAssertion(3),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
