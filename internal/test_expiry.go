package internal

import (
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// Tests Expiry
func testExpiry(stageHarness *test_case_harness.TestCaseHarness) error {
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

	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "set",
		Args:      []string{key, value, "px", "100"},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := setCommandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Successf("Received OK at %s", time.Now().Format("15:04:05.000"))
	logger.Infof("Fetching key %q at %s (should not be expired)", key, time.Now().Format("15:04:05.000"))

	getCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "get",
		Args:      []string{key},
		Assertion: resp_assertions.NewStringAssertion(value),
	}

	if err := getCommandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Debugf("Sleeping for 101ms")
	time.Sleep(101 * time.Millisecond)

	logger.Infof("Fetching key %q at %s (should be expired)", key, time.Now().Format("15:04:05.000"))

	getCommandTestCase = test_cases.SendCommandTestCase{
		Command:   "get",
		Args:      []string{key},
		Assertion: resp_assertions.NewNilAssertion(),
	}

	if err := getCommandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()
	return nil
}
