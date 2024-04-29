package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testRdbReadKey(stageHarness *test_case_harness.TestCaseHarness) error {
	RDBFileCreator, err := NewRDBFileCreator(stageHarness)
	if err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	defer RDBFileCreator.Cleanup()

	randomKeyAndValue := testerutils_random.RandomWords(2)
	randomKey := randomKeyAndValue[0]
	randomValue := randomKeyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: randomKey, value: randomValue}}); err != nil {
		return fmt.Errorf("CodeCrafters Tester Error: %s", err)
	}

	logger := stageHarness.Logger
	logger.Infof("Created RDB file with single key: %q", randomKey)

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run("--dir", RDBFileCreator.Dir,
		"--dbfilename", RDBFileCreator.Filename); err != nil {
		return err
	}

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client")
	if err != nil {
		return err
	}
	defer client.Close()

	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "KEYS",
		Args:                      []string{"*"},
		Assertion:                 resp_assertions.NewCommandAssertion(randomKey),
		ShouldSkipUnreadDataCheck: false,
	}

	return commandTestCase.Run(client, logger)
}
