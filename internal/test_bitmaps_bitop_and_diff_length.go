package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsBitopAndDiffLength(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keys := testerutils_random.RandomWords(3)
	destKey, longerKey, shorterKey := keys[0], keys[1], keys[2]
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	firstByteOffsets := testerutils_random.ShuffleArray([]int{0, 1, 2, 3, 4, 5, 6, 7})
	secondByteOffsets := testerutils_random.ShuffleArray([]int{8, 9, 10, 11, 12, 13, 14, 15})

	sharedCount := testerutils_random.RandomInt(2, 4)
	shorterOnlyCount := testerutils_random.RandomInt(1, 3)
	longerOnlyCount := testerutils_random.RandomInt(1, 3)

	sharedOffsets := firstByteOffsets[:sharedCount]
	remaining := firstByteOffsets[sharedCount:]
	shorterOnlyOffsets := remaining[:shorterOnlyCount]
	longerOnlyFirstByteOffsets := remaining[shorterOnlyCount : shorterOnlyCount+longerOnlyCount]
	longerSecondByteOffsets := secondByteOffsets[:testerutils_random.RandomInt(1, 3)]

	commands := make([]test_cases.CommandWithAssertion, 0, 32)
	for _, offset := range sharedOffsets {
		commands = append(commands, bitmapSetbitCommand(longerKey, offset), bitmapSetbitCommand(shorterKey, offset))
	}
	for _, offset := range shorterOnlyOffsets {
		commands = append(commands, bitmapSetbitCommand(shorterKey, offset))
	}
	for _, offset := range longerOnlyFirstByteOffsets {
		commands = append(commands, bitmapSetbitCommand(longerKey, offset))
	}
	for _, offset := range longerSecondByteOffsets {
		commands = append(commands, bitmapSetbitCommand(longerKey, offset))
	}

	sourceKeys := []string{longerKey, shorterKey}
	if testerutils_random.RandomInt(0, 2) == 1 {
		sourceKeys[0], sourceKeys[1] = sourceKeys[1], sourceKeys[0]
	}

	commands = append(commands, test_cases.CommandWithAssertion{
		Command:   []string{"BITOP", "AND", destKey, sourceKeys[0], sourceKeys[1]},
		Assertion: resp_assertions.NewIntegerAssertion(2),
	})
	for _, offset := range sharedOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 1))
	}
	for _, offset := range shorterOnlyOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 0))
	}
	for _, offset := range longerOnlyFirstByteOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 0))
	}
	for _, offset := range longerSecondByteOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 0))
	}

	missingDestKey := destKey + "_missing"
	commands = append(commands, test_cases.CommandWithAssertion{
		Command:   []string{"BITOP", "AND", missingDestKey, longerKey, missingKey},
		Assertion: resp_assertions.NewIntegerAssertion(2),
	})
	for _, offset := range append(append([]int{}, sharedOffsets...), longerSecondByteOffsets...) {
		commands = append(commands, bitmapGetbitCommand(missingDestKey, offset, 0))
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: commands,
	}

	return multiCommandTestCase.RunAll(client, logger)
}
