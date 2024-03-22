package test_cases

import (
	"fmt"
	"strings"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ReceiveCommandTestCase struct {
	Response                  resp_value.Value
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool
	ReceivedValue             resp_value.Value
}

func (t *ReceiveCommandTestCase) Run(conn *resp_connection.RespConnection, logger *logger.Logger) error {
	value, err := conn.ReadValue()
	if err != nil {
		return err
	}

	t.ReceivedValue = value

	result := t.Assertion.Run(value)
	if result.IsFailure() {
		return fmt.Errorf(strings.Join(result.ErrorMessages, "\n"))
	}

	logger.Successf(strings.Join(result.SuccessMessages, "\n"))

	if !t.ShouldSkipUnreadDataCheck {
		conn.ReadIntoBuffer() // Let's make sure there's no extra data

		if conn.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(conn.LastValueBytes)+conn.UnreadBuffer.String())
		}
	}

	if err := conn.SendValue(t.Response); err != nil {
		return err
	}

	return nil
}
