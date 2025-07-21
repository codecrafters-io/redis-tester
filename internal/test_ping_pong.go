package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPingPongOnce(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}

	logger.Debugln("Connection established, sending ping command...")

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}

func testPingPongMultiple(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client); err != nil {
			return err
		}
	}

	client.GetLogger().Debugf("Success, closing connection...")
	client.Close()

	return nil
}

func testPingPongConcurrent(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-2")
	if err != nil {
		return err
	}

	if err := runPing(logger, client2); err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}
	if err := runPing(logger, client1); err != nil {
		return err
	}
	if err := runPing(logger, client2); err != nil {
		return err
	}

	client1.GetLogger().Debugf("Success, closing connection...")
	client1.Close()

	client3, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client-3")
	if err != nil {
		return err
	}

	if err := runPing(logger, client3); err != nil {
		return err
	}

	client2.GetLogger().Debugf("Success, closing connection...")
	client2.Close()

	client3.GetLogger().Debugf("Success, closing connection...")
	client3.Close()

	return nil
}

func runPing(logger *logger.Logger, client *instrumented_resp_connection.InstrumentedRespConnection) error {
	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}
