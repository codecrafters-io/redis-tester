package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubSubscribe3(stageHarness *test_case_harness.TestCaseHarness) error {
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

	channels := random.RandomWords(2)
	subscribeTestCase1 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[0]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[0], 1),
	}
	if err := subscribeTestCase1.Run(client, logger); err != nil {
		return err
	}

	/* Test against ECHO/SET/GET (Unallowed commands) */

	/* SET */
	keyAndValue := random.RandomWords(2)
	setTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      keyAndValue,
		Assertion: resp_assertions.NewRegexErrorAssertion("^ERR (?i:Can't execute 'set').*"),
	}
	if err := setTestCase.Run(client, logger); err != nil {
		return err
	}

	/* GET */
	getTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      keyAndValue[1:],
		Assertion: resp_assertions.NewRegexErrorAssertion("^ERR (?i:Can't execute 'get').*"),
	}
	if err := getTestCase.Run(client, logger); err != nil {
		return err
	}

	/* ECHO */
	echoTestCase := test_cases.SendCommandTestCase{
		Command:   "ECHO",
		Args:      keyAndValue[1:],
		Assertion: resp_assertions.NewRegexErrorAssertion("^ERR (?i:Can't execute 'echo').*"),
	}
	if err := echoTestCase.Run(client, logger); err != nil {
		return err
	}

	subscribeTestCase2 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[1]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[1], 2),
	}
	return subscribeTestCase2.Run(client, logger)
}
