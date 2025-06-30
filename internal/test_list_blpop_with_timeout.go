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

func testListBlpopWithTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	if err := testWithTimeout(stageHarness); err != nil {
		return err
	}
	return testPushBeforeTimeout(stageHarness)
}

func testWithTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	randomListKey := testerutils_random.RandomWord()
	timeoutMS := testerutils_random.RandomInt(100, 500)
	timeoutArg := fmt.Sprintf("%.1f", float32(timeoutMS)/1000.0)
	timeoutDuration := time.Millisecond * time.Duration(timeoutMS)

	blockingTestCase := test_cases.NewBlockingCommandTestCase(
		&test_cases.SendCommandTestCase{
			Command:   "BLPOP",
			Args:      []string{randomListKey, timeoutArg},
			Assertion: resp_assertions.NewNilAssertion(),
		},
		&timeoutDuration,
	)
	blockingTestCase.Run(client, logger)
	return blockingTestCase.WaitForResult()
}

func testPushBeforeTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
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

	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()
	timeoutMS := testerutils_random.RandomInt(100, 500)
	timeoutArg := fmt.Sprintf("%.1f", float32(timeoutMS)/1000)

	timeout := time.Millisecond * time.Duration(timeoutMS)
	blockingTestCase := test_cases.NewBlockingCommandTestCase(
		&test_cases.SendCommandTestCase{
			Command:   "BLPOP",
			Args:      []string{listKey, timeoutArg},
			Assertion: resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue}),
		},
		&timeout,
	)
	blockingTestCase.Run(client1, logger)

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client2.Close()

	rpushTestCase := test_cases.SendCommandTestCase{
		Command:   "RPUSH",
		Args:      []string{listKey, pushValue},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}
	if err := rpushTestCase.Run(client2, logger); err != nil {
		return err
	}

	return blockingTestCase.ResumeAndWaitForResult()
}
