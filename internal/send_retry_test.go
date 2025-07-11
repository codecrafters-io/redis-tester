package internal

import (
	"testing"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/executable"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func TestSendRetry(t *testing.T) {
	h := &test_case_harness.TestCaseHarness{
		Logger:     logger.GetQuietLogger(""),
		Executable: executable.NewExecutable("./test_helpers/scenarios/send-command-retry/spawn_redis_server.sh"),
	}
	b := redis_executable.NewRedisExecutable(h)

	err := b.Run()
	if err != nil {
		t.Fatalf("Failed to run redis executable. Error: %s", err)
	}

	client, err := instrumented_resp_connection.NewFromAddr(h.Logger, "localhost:6379", "client")
	if err != nil {
		t.Fatalf("Failed to create client. Error: %s", err)
	}

	sendCommandTestCase := test_cases.SendCommandTestCase{
		Command:   "PING",
		Args:      nil,
		Assertion: resp_assertions.NewStringAssertion("PONG"),
		Retries:   1,
		ShouldRetryFunc: func(v resp_value.Value) bool {
			// retry if response is nil
			return resp_assertions.NewNilAssertion().Run(v) == nil
		},
	}

	err = sendCommandTestCase.Run(client, h.Logger)
	if err != nil {
		t.Fatalf("sendCommandTestCase.Run returned error: %s", err)
	}
}
