package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testdefaultUserAuthentication(stageHarness *test_case_harness.TestCaseHarness) error {
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

	aclSetUserTestCase := test_cases.AclSetuserTestCase{
		Username:  "default",
		Passwords: []string{password},
	}

	if err := aclSetUserTestCase.Run(firstClient, logger); err != nil {
		return err
	}

	// Run ACL WHOAMI from the existing client
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

	whoamiNoauthTestCase := test_cases.SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"WHOAMI"},
		Assertion: resp_assertions.NewRegexErrorAssertion("^NOAUTH.*"),
	}

	return whoamiNoauthTestCase.Run(secondClient, logger)
}
