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

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client, err := clientsSpawner.SpawnClientWithPrefix("client")
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

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client, err := clientsSpawner.SpawnClientWithPrefix("client-1")
	if err != nil {
		return err
	}

	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client); err != nil {
			return err
		}
	}

	client.GetLogger().Debugf("Success, closing connection...")

	return nil
}

func testPingPongConcurrent(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client1, err := clientsSpawner.SpawnNextClient()
	if err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}

	client2, err := clientsSpawner.SpawnNextClient()
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

	client3, err := clientsSpawner.SpawnNextClient()
	if err != nil {
		return err
	}

	if err := runPing(logger, client3); err != nil {
		return err
	}

	client2.GetLogger().Debugf("Success, closing connection...")

	client3.GetLogger().Debugf("Success, closing connection...")

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
