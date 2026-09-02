package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsBitopAnd(stageHarness *test_case_harness.TestCaseHarness) error {
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
	destKey, sourceKey1, sourceKey2 := keys[0], keys[1], keys[2]

	firstByteOffsets := testerutils_random.ShuffleArray([]int{0, 1, 2, 3, 4, 5, 6, 7})
	secondByteOffsets := testerutils_random.ShuffleArray([]int{8, 9, 10, 11, 12, 13, 14, 15})
	remaining := testerutils_random.ShuffleArray(append(firstByteOffsets[1:], secondByteOffsets[1:]...))

	sharedCount := testerutils_random.RandomInt(2, 4)
	source1OnlyCount := testerutils_random.RandomInt(1, 3)
	source2OnlyCount := testerutils_random.RandomInt(1, 3)

	sharedOffsets := []int{firstByteOffsets[0], secondByteOffsets[0]}
	sharedOffsets = append(sharedOffsets, remaining[:sharedCount-2]...)
	remaining = remaining[sharedCount-2:]
	source1OnlyOffsets := remaining[:source1OnlyCount]
	source2OnlyOffsets := remaining[source1OnlyCount : source1OnlyCount+source2OnlyCount]

	commands := make([]test_cases.CommandWithAssertion, 0, 2*len(sharedOffsets)+len(source1OnlyOffsets)+len(source2OnlyOffsets)+1+len(sharedOffsets)+len(source1OnlyOffsets)+len(source2OnlyOffsets))
	for _, offset := range sharedOffsets {
		commands = append(commands, bitmapSetbitCommand(sourceKey1, offset), bitmapSetbitCommand(sourceKey2, offset))
	}
	for _, offset := range source1OnlyOffsets {
		commands = append(commands, bitmapSetbitCommand(sourceKey1, offset))
	}
	for _, offset := range source2OnlyOffsets {
		commands = append(commands, bitmapSetbitCommand(sourceKey2, offset))
	}

	commands = append(commands, test_cases.CommandWithAssertion{
		Command:   []string{"BITOP", "AND", destKey, sourceKey1, sourceKey2},
		Assertion: resp_assertions.NewIntegerAssertion(2),
	})
	for _, offset := range sharedOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 1))
	}
	for _, offset := range source1OnlyOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 0))
	}
	for _, offset := range source2OnlyOffsets {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 0))
	}

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: commands,
	}

	return multiCommandTestCase.RunAll(client, logger)
}

func bitmapSetbitCommand(key string, offset int) test_cases.CommandWithAssertion {
	return test_cases.CommandWithAssertion{
		Command:   []string{"SETBIT", key, strconv.Itoa(offset), "1"},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}
}

func bitmapGetbitCommand(key string, offset int, expected int) test_cases.CommandWithAssertion {
	return test_cases.CommandWithAssertion{
		Command:   []string{"GETBIT", key, strconv.Itoa(offset)},
		Assertion: resp_assertions.NewIntegerAssertion(expected),
	}
}
