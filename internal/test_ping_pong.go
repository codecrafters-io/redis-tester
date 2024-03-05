package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_client"
	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/client"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPingPongOnce(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "")
	if err != nil {
		return err
	}

	if err := runPing(logger, client); err != nil {
		return err
	}

	logger.Debugln("Connection established, sending ping command...")

	commandTestCase := test_cases.CommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewStringAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}

func testPingPongMultiple(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client); err != nil {
			return err
		}
	}

	logger.Debugf("Success, closing connection...")
	client.Close()

	return nil
}

func testPingPongConcurrent(stageHarness *test_case_harness.TestCaseHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client1, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client-1")
	if err != nil {
		return err
	}

	if err := runPing(logger, client1); err != nil {
		return err
	}

	client2, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client-2")
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

	logger.Debugf("client-%d: Success, closing connection...", 1)
	client1.Close()

	client3, err := instrumented_resp_client.NewInstrumentedRespClient(stageHarness, "localhost:6379", "client-3")
	if err != nil {
		return err
	}

	if err := runPing(logger, client3); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 2)
	client2.Close()
	logger.Debugf("client-%d: Success, closing connection...", 3)
	client3.Close()

	return nil
}

func runPing(logger *logger.Logger, client *resp_client.RespClient) error {
	// // TODO: Test concurrent network activity
	// for i := 1; i <= 20; i++ {
	// 	logger.Infof("Connecting to port 6900...")
	// 	l, err := net.Listen("tcp", "0.0.0.0:6900")
	// 	if err != nil {
	// 		fmt.Println("Failed to bind to port 6900")
	// 		return err
	// 	}

	// 	logger.Infof("Connected to port 6900.")
	// 	time.Sleep(100 * time.Millisecond)

	// 	l.Close()
	// 	logger.Infof("Removed listener on port 6900.")
	// }

	commandTestCase := test_cases.CommandTestCase{
		Command:   "ping",
		Args:      []string{},
		Assertion: resp_assertions.NewStringAssertion("PONG"),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	return nil
}
