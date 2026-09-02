package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBitmapsBitopOr(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keys := testerutils_random.RandomWords(6)
	sameDest, sameKey1, sameKey2 := keys[0], keys[1], keys[2]
	diffDest, longerKey, shorterKey := keys[3], keys[4], keys[5]

	commands := make([]test_cases.CommandWithAssertion, 0, 64)
	commands = append(commands, bitopOrSameLengthCommands(sameDest, sameKey1, sameKey2)...)
	commands = append(commands, bitopOrDiffLengthCommands(diffDest, longerKey, shorterKey)...)

	return test_cases.MultiCommandTestCase{CommandWithAssertions: commands}.RunAll(client, logger)
}

func bitopOrSameLengthCommands(destKey, sourceKey1, sourceKey2 string) []test_cases.CommandWithAssertion {
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
	unsetOffset := remaining[source1OnlyCount+source2OnlyCount]

	var commands []test_cases.CommandWithAssertion
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
		Command:   append([]string{"BITOP", "OR", destKey}, bitmapShuffledSourceKeys(sourceKey1, sourceKey2)...),
		Assertion: resp_assertions.NewIntegerAssertion(2),
	})
	for _, offset := range append(append(sharedOffsets, source1OnlyOffsets...), source2OnlyOffsets...) {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 1))
	}
	commands = append(commands, bitmapGetbitCommand(destKey, unsetOffset, 0))

	return commands
}

func bitopOrDiffLengthCommands(destKey, longerKey, shorterKey string) []test_cases.CommandWithAssertion {
	firstByteOffsets := testerutils_random.ShuffleArray([]int{0, 1, 2, 3, 4, 5, 6, 7})
	secondByteOffsets := testerutils_random.ShuffleArray([]int{8, 9, 10, 11, 12, 13, 14, 15})

	sharedCount := testerutils_random.RandomInt(2, 4)
	shorterOnlyCount := testerutils_random.RandomInt(1, 3)
	longerOnlyCount := testerutils_random.RandomInt(1, 3)
	longerSecondByteCount := testerutils_random.RandomInt(1, 3)

	sharedOffsets := firstByteOffsets[:sharedCount]
	remaining := firstByteOffsets[sharedCount:]
	shorterOnlyOffsets := remaining[:shorterOnlyCount]
	longerOnlyFirstByteOffsets := remaining[shorterOnlyCount : shorterOnlyCount+longerOnlyCount]
	longerSecondByteOffsets := secondByteOffsets[:longerSecondByteCount]
	unsetSecondByteOffset := secondByteOffsets[longerSecondByteCount]

	var commands []test_cases.CommandWithAssertion
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

	commands = append(commands, test_cases.CommandWithAssertion{
		Command:   append([]string{"BITOP", "OR", destKey}, bitmapShuffledSourceKeys(longerKey, shorterKey)...),
		Assertion: resp_assertions.NewIntegerAssertion(2),
	})
	for _, offset := range append(append(append(sharedOffsets, shorterOnlyOffsets...), longerOnlyFirstByteOffsets...), longerSecondByteOffsets...) {
		commands = append(commands, bitmapGetbitCommand(destKey, offset, 1))
	}
	commands = append(commands, bitmapGetbitCommand(destKey, unsetSecondByteOffset, 0))

	return commands
}

func bitmapShuffledSourceKeys(key1, key2 string) []string {
	keys := []string{key1, key2}
	if testerutils_random.RandomInt(0, 2) == 1 {
		keys[0], keys[1] = keys[1], keys[0]
	}
	return keys
}
