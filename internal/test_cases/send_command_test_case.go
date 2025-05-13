package test_cases

import (
	"fmt"
	"strings"
	"sync"
	"time"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
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

func (t *SendCommandTestCase) Run(client *resp_client.RespConnection, logger *logger.Logger) error {
	var value resp_value.Value
	var err error

	for attempt := 0; attempt <= t.Retries; attempt++ {
		if attempt > 0 {
			logger.Infof("Retrying... (%d/%d attempts)", attempt, t.Retries)
		}

		command := strings.ToUpper(t.Command)

		if err = client.SendCommand(command, t.Args...); err != nil {
			return err
		}

		t.readMutex.Lock()
		defer t.readMutex.Unlock()

		value, err = client.ReadValue()
		if err != nil {
			return err
		}

		if t.Retries == 0 {
			break
		}

		if t.ShouldRetryFunc == nil {
			panic(fmt.Sprintf("Received SendCommand with retries: %d but no ShouldRetryFunc.", t.Retries))
		} else {
			if t.ShouldRetryFunc(value) {
				// If ShouldRetryFunc returns true, we sleep and retry.
				time.Sleep(500 * time.Millisecond)
			} else {
				break
			}
		}
	}

	t.ReceivedResponse = value

	if err = t.Assertion.Run(value); err != nil {
		return err
	}

	if !t.ShouldSkipUnreadDataCheck {
		client.ReadIntoBuffer() // Let's make sure there's no extra data

		if client.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", client.UnreadBuffer.String())
		}
	}

	logger.Successf("Received %s", value.FormattedString())
	return nil
}

func (t *SendCommandTestCase) PauseReadingResponse() {
	t.readMutex.Lock()
}

func (t *SendCommandTestCase) ResumeReadingResponse() {
	t.readMutex.Unlock()
}
