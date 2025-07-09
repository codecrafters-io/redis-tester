package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// Tests 'GET, SET'
func testGetSet(stageHarness *test_case_harness.TestCaseHarness) error {
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

	keyAndValue := random.RandomWords(2)

	key := keyAndValue[0]
	value := keyAndValue[1]

	logger.Debugf("Setting key %s to %s", key, value)
	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "set",
		Args:      []string{key, value},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Debugf("Getting key %s", key)

	getCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "get",
		Args:      []string{key},
		Assertion: resp_assertions.NewStringAssertion(value),
	}

	if err := getCommandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()
	return nil
}
