package test_cases

import (
	"fmt"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

type ReceiveCommandTestCase struct {
	Response                  resp_value.Value
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool
	ReceivedCommand           resp_value.Value
}

func (t *ReceiveCommandTestCase) Run(conn *resp_connection.RespConnection, logger *logger.Logger) error {
	value, err := conn.ReadValue()
	if err != nil {
		return err
	}

	t.ReceivedCommand = value

	if err = t.Assertion.Run(value); err != nil {
		return err
	}

	logger.Successf("Received %s", value.FormattedString())

	if !t.ShouldSkipUnreadDataCheck {
		conn.ReadIntoBuffer() // Let's make sure there's no extra data

		if conn.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(conn.LastValueBytes)+conn.UnreadBuffer.String())
		}
	}

	if err := conn.SendCommand(t.Response); err != nil {
		return err
	}

	return nil
}
