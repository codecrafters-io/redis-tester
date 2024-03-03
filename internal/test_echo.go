package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_client"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// Tests 'ECHO'
func testEcho(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "")
	if err != nil {
		return err
	}

	randomWord := random.RandomWord()

	commandTestCase := test_cases.CommandTestCase{
		Command:   "echo",
		Args:      []string{randomWord},
		Assertion: resp_assertions.NewStringAssertion(randomWord),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()

	return nil
}
