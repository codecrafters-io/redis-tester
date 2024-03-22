package test_cases

import (
	"fmt"
	"strings"
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
}

func (t SendCommandTestCase) Run(client *resp_client.RespConnection, logger *logger.Logger) error {
	var value resp_value.Value
	var err error

	for attempt := 0; attempt <= t.Retries; attempt++ {
		if attempt > 0 {
			logger.Debugf("(Attempt %d/%d)", attempt, t.Retries)
		}

		if err = client.SendCommand(t.Command, t.Args...); err != nil {
			return err
		}

		value, err = client.ReadValue()
		if err != nil {
			return err
		}

		result := resp_assertions.NewNilAssertion().Run(value)
		if result.IsFailure() {
			// value is not nil, we can stop retrying
			break
		} else {
			// value is nil, sleep then retry
			time.Sleep(500 * time.Millisecond)
		}
	}

	result := t.Assertion.Run(value)
	if result.IsFailure() {
		if result.SuccessMessages != nil {
			logger.Successf(strings.Join(result.SuccessMessages, "\n"))
		}
		return fmt.Errorf(strings.Join(result.ErrorMessages, "\n"))
	}

	logger.Successf(strings.Join(result.SuccessMessages, "\n"))

	if !t.ShouldSkipUnreadDataCheck {
		client.ReadIntoBuffer() // Let's make sure there's no extra data

		if client.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(client.LastValueBytes)+client.UnreadBuffer.String())
		}
	}

	return nil
}
