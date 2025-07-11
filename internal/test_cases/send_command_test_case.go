package test_cases

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type SendCommandTestCase struct {
	Command                   string
	Args                      []string
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool
	Retries                   int
	ShouldRetryFunc           func(resp_value.Value) bool

	// ReceivedResponse is set after the test case is run
	ReceivedResponse resp_value.Value

	readMutex sync.Mutex
}

func (t *SendCommandTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	receiveValueTestCase := ReceiveValueTestCase{
		Assertion:                 t.Assertion,
		ShouldSkipUnreadDataCheck: t.ShouldSkipUnreadDataCheck,
	}

	for attempt := 0; attempt <= t.Retries; attempt++ {
		if attempt > 0 {
			logger.Infof("Retrying... (%d/%d attempts)", attempt, t.Retries)
		}

		command := strings.ToUpper(t.Command)

		if err := client.SendCommand(command, t.Args...); err != nil {
			return err
		}

		t.readMutex.Lock()
		err := receiveValueTestCase.RunWithoutAssert(client)
		t.readMutex.Unlock()

		if err != nil {
			return err
		}

		if t.Retries == 0 {
			break
		}

		if t.ShouldRetryFunc == nil {
			panic(fmt.Sprintf("Received SendCommand with retries: %d but no ShouldRetryFunc.", t.Retries))
		} else {
			if t.ShouldRetryFunc(receiveValueTestCase.ActualValue) {
				// If ShouldRetryFunc returns true, we sleep and retry.
				time.Sleep(500 * time.Millisecond)
			} else {
				break
			}
		}
	}
	t.ReceivedResponse = receiveValueTestCase.ActualValue

	return receiveValueTestCase.Assert(client, logger)
}

func (t *SendCommandTestCase) PauseReadingResponse() {
	t.readMutex.Lock()
}

func (t *SendCommandTestCase) ResumeReadingResponse() {
	t.readMutex.Unlock()
}
