package internal

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsBitcount(stageHarness *test_case_harness.TestCaseHarness) error {
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
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	byteCount := testerutils_random.RandomInt(3, 5)
	popcount := make([]int, byteCount)
	var offsets []int
	for byteIndex := range byteCount {
		bitsInByte := testerutils_random.RandomInt(1, 3)
		bitIndexes := testerutils_random.ShuffleArray([]int{0, 1, 2, 3, 4, 5, 6, 7})[:bitsInByte]
		for _, bitIndex := range bitIndexes {
			offsets = append(offsets, byteIndex*8+bitIndex)
		}
		popcount[byteIndex] = bitsInByte
	}

	middle := testerutils_random.RandomInt(1, byteCount-1)
	last := byteCount - 1
	totalCount := redisBitcountInByteRange(popcount, 0, last)

	commands := make([]test_cases.CommandWithAssertion, 0, len(offsets)+8)
	for _, offset := range offsets {
		commands = append(commands, test_cases.CommandWithAssertion{
			Command:   []string{"SETBIT", key, strconv.Itoa(offset), "1"},
			Assertion: resp_assertions.NewIntegerAssertion(0),
		})
	}

	commands = append(commands,
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key},
			Assertion: resp_assertions.NewIntegerAssertion(totalCount),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, "0", strconv.Itoa(middle)},
			Assertion: resp_assertions.NewIntegerAssertion(redisBitcountInByteRange(popcount, 0, middle)),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, strconv.Itoa(middle), strconv.Itoa(last)},
			Assertion: resp_assertions.NewIntegerAssertion(redisBitcountInByteRange(popcount, middle, last)),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, "0", strconv.Itoa(last)},
			Assertion: resp_assertions.NewIntegerAssertion(totalCount),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, "1", "0"},
			Assertion: resp_assertions.NewIntegerAssertion(0),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, "0", strconv.Itoa(byteCount * 2)},
			Assertion: resp_assertions.NewIntegerAssertion(totalCount),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", key, strconv.Itoa(byteCount), strconv.Itoa(byteCount)},
			Assertion: resp_assertions.NewIntegerAssertion(0),
		},
		test_cases.CommandWithAssertion{
			Command:   []string{"BITCOUNT", missingKey},
			Assertion: resp_assertions.NewIntegerAssertion(0),
		},
	)

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: commands,
	}

	return multiCommandTestCase.RunAll(client, logger)
}
