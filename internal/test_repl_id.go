package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplReplicationID(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run([]string{}); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	defer client.Close()

	commandTestCase := test_cases.CommandTestCase{
		Command:                   "INFO",
		Args:                      []string{"replication"},
		Assertion:                 resp_assertions.NewRegexStringAssertion("master_replid:([a-zA-Z0-9]+)[\\s\\S]*master_repl_offset:0"),
		ShouldSkipUnreadDataCheck: true,
	}
	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
