package internal

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testListBlpopNoTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client1.Close()
	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client2.Close()

	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()

	assertion := resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue})
	blocker1 := test_cases.NewBlockingCommandTestCase(
		&test_cases.SendCommandTestCase{
			Command:   "BLPOP",
			Args:      []string{listKey, "0"},
			Assertion: assertion,
		},
		nil,
	)
	blocker2 := test_cases.NewBlockingCommandTestCase(
		&test_cases.SendCommandTestCase{
			Command:   "BLPOP",
			Args:      []string{listKey, "0"},
			Assertion: assertion,
		},
		nil,
	)

	blocker1.Run(client1, logger)
	blocker2.Run(client2, logger)

	client3, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-3")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client3.Close()

	rpushTestCase := test_cases.SendCommandTestCase{
		Command:   "RPUSH",
		Args:      []string{listKey, pushValue},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}
	if err := rpushTestCase.Run(client3, logger); err != nil {
		return err
	}

	blocker1.Resume()
	blocker2.Resume()

	if err := blocker1.WaitForResult(); err != nil {
		return err
	}

	// check if the server responds to client-2 (it shouldn't)
	blocked := make(chan error, 1)
	go func() {
		blocked <- blocker2.WaitForResult()
	}()
	select {
	case <-blocked:
		return fmt.Errorf("Server responded to %s for BLPOP", client2.GetIdentifier())
	case <-time.After(10 * time.Millisecond):
	}
	return nil
}
