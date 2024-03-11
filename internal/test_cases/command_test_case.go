package test_cases

import (
	"fmt"
	"time"

	resp_utils "github.com/codecrafters-io/redis-tester/internal/resp"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

type CommandTestCase struct {
	Command                   string
	Args                      []string
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool
	ShouldRetry               bool

	Response resp_value.Value
	Offset   int
}

func (t *CommandTestCase) Run(client *resp_connection.RespConnection, logger *logger.Logger) error {
	maxRetries := 0
	var value resp_value.Value
	var err error
	if t.ShouldRetry {
		maxRetries = 4
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debugf("(Attempt %d/%d)", attempt+1, maxRetries+1)
		}
		respValue := resp_value.NewStringArrayValue(append([]string{t.Command}, t.Args...))
		if err = client.SendCommand(respValue); err != nil {
			return err
		}

		value, err = client.ReadValue()
		if err != nil {
			return err
		}
		// ToDo Should use NilAssertion ?
		if value.Type != "NIL" {
			break
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if err := t.Assertion.Run(value); err != nil {
		return err
	}

	t.Response = value
	if t.Response.Type == resp_value.ARRAY {
		t.Offset = resp_utils.GetByteOffsetHelper(t.Response.FormattedString())
	}

	if !t.ShouldSkipUnreadDataCheck {
		client.ReadIntoBuffer() // Let's make sure there's no extra data

		if client.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(client.LastValueBytes)+client.UnreadBuffer.String())
		}
	}

	logger.Successf("Received %s", value.FormattedString())
	return nil
}
