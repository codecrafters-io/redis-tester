package test_cases

import (
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// Example usage of MultiClientBlockingCommandTestCase
func ExampleMultiClientBlockingCommandTestCase(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// Create multiple blocking clients
	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		return err
	}
	defer client1.Close()

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		return err
	}
	defer client2.Close()

	client3, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-3")
	if err != nil {
		return err
	}
	defer client3.Close()

	// Create unlocker client
	unlockerClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "unlocker")
	if err != nil {
		return err
	}
	defer unlockerClient.Close()

	// Setup test data
	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()

	// Create the multi-client blocking test case
	// This will send BLPOP from all 3 clients, then RPUSH from unlocker client
	// Only clients 0 and 1 should respond (get the value), client 2 should not respond
	multiBlockingTest := NewMultiClientBlockingCommandTestCase(
		"BLPOP",                // blocking command
		[]string{listKey, "0"}, // blocking args
		resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue}), // blocking assertion
		"RPUSH",                                // unlocker command
		[]string{listKey, pushValue},           // unlocker args
		resp_assertions.NewIntegerAssertion(1), // unlocker assertion
		[]*instrumented_resp_connection.InstrumentedRespConnection{client1, client2, client3}, // blocking clients
		unlockerClient, // unlocker client
		[]int{0, 1},    // only clients 0 and 1 should respond
		nil,            // use default timeout
	)

	// Run the complete test case
	return multiBlockingTest.RunComplete(logger)
}

// Example with custom timeout
func ExampleMultiClientBlockingWithTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// Create clients
	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		return err
	}
	defer client1.Close()

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		return err
	}
	defer client2.Close()

	unlockerClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "unlocker")
	if err != nil {
		return err
	}
	defer unlockerClient.Close()

	// Setup test data
	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()
	timeout := 5 * time.Second

	// Create test case with custom timeout
	multiBlockingTest := NewMultiClientBlockingCommandTestCase(
		"BLPOP",
		[]string{listKey, "0"},
		resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue}),
		"RPUSH",
		[]string{listKey, pushValue},
		resp_assertions.NewIntegerAssertion(1),
		[]*instrumented_resp_connection.InstrumentedRespConnection{client1, client2},
		unlockerClient,
		[]int{0}, // only client 0 should respond
		&timeout,
	)

	// Run the complete test case
	return multiBlockingTest.RunComplete(logger)
}

// Example with step-by-step execution
func ExampleMultiClientBlockingStepByStep(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// Create clients
	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		return err
	}
	defer client1.Close()

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		return err
	}
	defer client2.Close()

	unlockerClient, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "unlocker")
	if err != nil {
		return err
	}
	defer unlockerClient.Close()

	// Setup test data
	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()

	// Create test case
	multiBlockingTest := NewMultiClientBlockingCommandTestCase(
		"BLPOP",
		[]string{listKey, "0"},
		resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue}),
		"RPUSH",
		[]string{listKey, pushValue},
		resp_assertions.NewIntegerAssertion(1),
		[]*instrumented_resp_connection.InstrumentedRespConnection{client1, client2},
		unlockerClient,
		[]int{0}, // only client 0 should respond
		nil,
	)

	// Step 1: Start blocking commands
	multiBlockingTest.Run(logger)

	// Step 2: Wait a bit for commands to start
	time.Sleep(100 * time.Millisecond)

	// Step 3: Send unlocker command
	if err := multiBlockingTest.SendUnlockerCommand(logger); err != nil {
		return err
	}

	// Step 4: Validate results
	return multiBlockingTest.ValidateResults(logger)
}
