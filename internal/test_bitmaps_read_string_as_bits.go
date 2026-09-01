package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReadStringAsBits(stageHarness *test_case_harness.TestCaseHarness) error {
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
	value := testerutils_random.RandomWord()

	oneOffsets := redisBitOffsets(value, 1)
	zeroOffsets := redisBitOffsets(value, 0)
	oneOffset := oneOffsets[testerutils_random.RandomInt(0, len(oneOffsets))]
	zeroOffset := zeroOffsets[testerutils_random.RandomInt(0, len(zeroOffsets))]
	unsetOffset := len(value)*8 + testerutils_random.RandomInt(0, 8)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key, value},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"GETBIT", key, strconv.Itoa(oneOffset)},
				Assertion: resp_assertions.NewIntegerAssertion(1),
			},
			{
				Command:   []string{"GETBIT", key, strconv.Itoa(zeroOffset)},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
			{
				Command:   []string{"GETBIT", key, strconv.Itoa(unsetOffset)},
				Assertion: resp_assertions.NewIntegerAssertion(0),
			},
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
