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
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}
	if err := testOnlyTimeout(stageHarness); err != nil {
		return err
	}
	return testPushBeforeTimeout(stageHarness)
}

func testOnlyTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	randomListKey := testerutils_random.RandomWord()
	timeoutMS := testerutils_random.RandomInt(1, 5) * 100
	timeoutArg := fmt.Sprintf("%.1f", float32(timeoutMS)/1000)
	timeoutDuration := time.Millisecond * time.Duration(timeoutMS)

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "BLPOP",
		Args:      []string{randomListKey, timeoutArg},
		Assertion: resp_assertions.NewNilArrayAssertion(),
	}

	resultChan := make(chan error, 1)
	start := time.Now()
	go func() {
		resultChan <- sendCommandTestCase.Run(client, logger)
	}()

	err = <-resultChan
	end := time.Now()
	if err != nil {
		return err
	}
	if end.Before(start.Add(timeoutDuration)) {
		return fmt.Errorf("%s received a response before timeout of %s", client.GetIdentifier(), timeoutDuration.String())
	}
	return nil
}

func testPushBeforeTimeout(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	listKey := testerutils_random.RandomWord()
	pushValue := testerutils_random.RandomWord()
	timeoutMS := testerutils_random.RandomInt(1, 5) * 100
	timeoutArg := fmt.Sprintf("%.1f", float32(timeoutMS)/1000)

	blPopResponseAssertion := resp_assertions.NewOrderedStringArrayAssertion([]string{listKey, pushValue})
	blockingClientGroupTestCase := test_cases.BlockingClientGroupTestCase{}
	blockingClientGroupTestCase.AddClientWithExpectedResponse(clients[0], "BLPOP", []string{listKey, timeoutArg}, blPopResponseAssertion)

	if err := blockingClientGroupTestCase.SendBlockingCommands(); err != nil {
		return err
	}

	rpushTestCase := test_cases.SendCommandTestCase{
		Command:   "RPUSH",
		Args:      []string{listKey, pushValue},
		Assertion: resp_assertions.NewIntegerAssertion(1),
	}
	if err := rpushTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	return blockingClientGroupTestCase.AssertResponses(logger)
}
