package internal

import (
	"fmt"
	"regexp"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplInfo(stageHarness *test_case_harness.TestCaseHarness) error {
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

	commandTestCase := test_cases.SendCommandTestCase{
		Command:                   "INFO",
		Args:                      []string{"replication"},
		Assertion:                 resp_assertions.NewNoopAssertion(),
		ShouldSkipUnreadDataCheck: true,
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	responseValue := commandTestCase.ReceivedResponse

	if responseValue.Type != resp_value.BULK_STRING && responseValue.Type != resp_value.SIMPLE_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", responseValue.Type)
	}

	var patternMatchError error

	if !regexp.MustCompile("role:").Match([]byte(responseValue.String())) {
		patternMatchError = fmt.Errorf("Expected role to be present in response. Got: %q", responseValue.String())
	}

	if regexp.MustCompile("role:master").Match([]byte(responseValue.String())) {
		logger.Successf("Found role:master in response.")
	} else {
		patternMatchError = fmt.Errorf("Expected role to be master in response. Got: %q", responseValue.String())
	}

	return patternMatchError
}
