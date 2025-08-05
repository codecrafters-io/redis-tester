package internal

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testZsetZcard(stageHarness *test_case_harness.TestCaseHarness) error {
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

	zsetKey := testerutils_random.RandomWord()
	zsetSize := testerutils_random.RandomInt(4, 8)
	members := GenerateRandomZSetMembers(ZsetMemberGenerationOption{
		Count: zsetSize,
	})

	zsetTestCase := test_cases.NewZsetTestCase(zsetKey)
	for _, m := range members {
		zsetTestCase.AddMember(m.Name, m.Score)
	}

	if err := zsetTestCase.RunZaddAll(client, logger); err != nil {
		return err
	}

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{zsetKey},
		Assertion: resp_assertions.NewIntegerAssertion(zsetSize),
	}

	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	/* Update an existing member */
	updateAtIdx := testerutils_random.RandomInt(0, zsetSize)
	newScore := strconv.FormatFloat(GetRandomZSetScore(), 'f', -1, 64)
	sendCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "ZADD",
		Args:      []string{zsetKey, newScore, members[updateAtIdx].Name},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}
	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	/* Check cardinality again: size shouldn't change */
	sendCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{zsetKey},
		Assertion: resp_assertions.NewIntegerAssertion(zsetSize),
	}

	if err := sendCommandTestCase.Run(client, logger); err != nil {
		return err
	}

	/* Check cardinality of missing key */
	missing_key := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))
	missingKeyZcardTestCase := test_cases.SendCommandTestCase{
		Command:   "ZCARD",
		Args:      []string{missing_key},
		Assertion: resp_assertions.NewIntegerAssertion(0),
	}

	return missingKeyZcardTestCase.Run(client, logger)
}
