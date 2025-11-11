package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testAuthCommandAuthentication(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	firstClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	defer firstClient.Close()

	// Set default user password
	password := fmt.Sprintf("%s-%d", random.RandomWord(), random.RandomInt(1, 1000))

	aclSetUserTestCase := test_cases.SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"SETUSER", "default", fmt.Sprintf(">%s", password)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := aclSetUserTestCase.Run(firstClient, logger); err != nil {
		return err
	}

	// Run ACL WHOAMI
	whoamiTestCase := test_cases.AclWhoamiTestCase{
		ExpectedUsername: "default",
	}

	if err := whoamiTestCase.Run(firstClient, logger); err != nil {
		return err
	}

	// Spawn a new client
	secondClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	defer secondClient.Close()

	// Run ACL WHOAMI without authentication
	whoamiNoauthTestCase := test_cases.AclWhoamiErrorTestCase{
		ExpectedErrorPattern: "^NOAUTH.*",
	}
	if err := whoamiNoauthTestCase.Run(secondClient, logger); err != nil {
		return err
	}

	// Authenticate as default user
	authTestCase := test_cases.SendCommandTestCase{
		Command:   "AUTH",
		Args:      []string{"default", password},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := authTestCase.Run(secondClient, logger); err != nil {
		return err
	}

	// Re-run ACL WHOAMI as the default user
	whoamiSuccessTestCase := test_cases.AclWhoamiTestCase{
		ExpectedUsername: "default",
	}

	if err := whoamiSuccessTestCase.Run(secondClient, logger); err != nil {
		return err
	}

	return testAuthenticatedClientUsingRandomCommands(secondClient, logger)
}

func testAuthenticatedClientUsingRandomCommands(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	argumentForEchoCommand := random.RandomWord()

	allTestCases := []test_cases.SendCommandTestCase{
		{
			Command:   "SET",
			Args:      []string{fmt.Sprintf("key-%d", random.RandomInt(1, 1000)), fmt.Sprintf("value-%d", random.RandomInt(1, 1000))},
			Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
		},
		{
			Command:   "GET",
			Args:      []string{fmt.Sprintf("nonexistent-key-%d", random.RandomInt(1, 1000))},
			Assertion: resp_assertions.NewNilAssertion(),
		},
		{
			Command:   "PING",
			Args:      []string{},
			Assertion: resp_assertions.NewSimpleStringAssertion("PONG"),
		},
		{
			Command:   "ECHO",
			Args:      []string{argumentForEchoCommand},
			Assertion: resp_assertions.NewBulkStringAssertion(argumentForEchoCommand),
		},
	}

	selectedTestCases := random.RandomElementsFromArray(allTestCases, 2)

	for i := range selectedTestCases {
		// var range copies lock: indexing avoids copying of lock inside test case
		if err := selectedTestCases[i].Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
