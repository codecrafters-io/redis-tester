package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsSet1(stageHarness *test_case_harness.TestCaseHarness) error {
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

	bitmapKey := testerutils_random.RandomWord()
	bitOffset := strconv.Itoa(testerutils_random.RandomInt(0, 8))

	// SETBIT returns the original bit at offset.
	// A missing key is treated as all zeros, so the first SETBIT returns 0.
	// Setting the same offset to 0 then returns 1 (the bit that was just set).
	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SETBIT", bitmapKey, bitOffset, "1"},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"SETBIT", bitmapKey, bitOffset, "0"},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
