package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGetUserResponse(stageHarness *test_case_harness.TestCaseHarness) error {
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

	aclGetUserTestCase := test_cases.AclGetuserTestCase{
		Username: "default",
	}

	return aclGetUserTestCase.RunForFlagsTemplateOnly(client, logger)
}
