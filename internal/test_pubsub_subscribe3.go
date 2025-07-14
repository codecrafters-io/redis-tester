package internal

import (
	"regexp"

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
	errRegex := regexp.MustCompile(`^ERR (?i:Can't execute 'set').*`)
	setTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      keyAndValue,
		Assertion: resp_assertions.NewErrorRegexAssertion(*errRegex),
	}
	if err := setTestCase.Run(client, logger); err != nil {
		return err
	}

	/* GET */
	errRegex = regexp.MustCompile(`^ERR (?i:Can't execute 'get').*`)
	getTestCase := test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      keyAndValue[1:],
		Assertion: resp_assertions.NewErrorRegexAssertion(*errRegex),
	}
	if err := getTestCase.Run(client, logger); err != nil {
		return err
	}

	/* ECHO */
	errRegex = regexp.MustCompile(`^ERR (?i:Can't execute 'echo').*`)
	echoTestCase := test_cases.SendCommandTestCase{
		Command:   "ECHO",
		Args:      keyAndValue[1:],
		Assertion: resp_assertions.NewErrorRegexAssertion(*errRegex),
	}
	if err := echoTestCase.Run(client, logger); err != nil {
		return err
	}

	/* Test against PING and SUBSCRIBE (Allowed commands) */
	pingTestCase := test_cases.SendCommandTestCase{
		Command:   "PING",
		Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{"pong", ""}),
	}
	if err := pingTestCase.Run(client, logger); err != nil {
		return err
	}

	subscribeTestCase2 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[1]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[1], 2),
	}

	return subscribeTestCase2.Run(client, logger)
}
