package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testSetUserPassword(stageHarness *test_case_harness.TestCaseHarness) error {
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

	// Run ACL GETUSER default
	aclGetUserTestCase := test_cases.AclGetuserTestCase{
		Username:                 "default",
		FlagsExpectedToBePresent: []string{"nopass"},
		ExpectedPasswords:        []string{},
	}

	if err := aclGetUserTestCase.Run(client, logger); err != nil {
		return err
	}

	// Set default user password
	password := fmt.Sprintf("%s-%d", random.RandomWord(), random.RandomInt(1, 1000))

	aclSetUserTestCase := test_cases.AclSetuserTestCase{
		Username:  "default",
		Passwords: []string{password},
	}

	if err := aclSetUserTestCase.Run(client, logger); err != nil {
		return err
	}

	// Again run ACL GETUSER default

	aclGetUserTestCase2 := test_cases.AclGetuserTestCase{
		Username:                "default",
		FlagsExpectedToBeAbsent: []string{"nopass"},
		ExpectedPasswords:       []string{password},
	}

	return aclGetUserTestCase2.Run(client, logger)
}
