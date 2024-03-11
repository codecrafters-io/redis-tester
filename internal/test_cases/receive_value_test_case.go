package test_cases

import (
	"fmt"

	resp_utils "github.com/codecrafters-io/redis-tester/internal/resp"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

type ReceiveValueTestCase struct {
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool

	// This is set after the test is run
	ActualValue resp_value.Value
	Offset      int
}

func (t *ReceiveValueTestCase) Run(conn *resp_connection.RespConnection, logger *logger.Logger) error {
	value, err := conn.ReadValue()
	if err != nil {
		return err
	}

	t.ActualValue = value
	if t.ActualValue.Type == resp_value.ARRAY {
		t.Offset = resp_utils.GetByteOffsetHelper(t.ActualValue.FormattedString())
	}

	if err = t.Assertion.Run(value); err != nil {
		return err
	}

	if !t.ShouldSkipUnreadDataCheck {
		conn.ReadIntoBuffer() // Let's make sure there's no extra data

		if conn.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(conn.LastValueBytes)+conn.UnreadBuffer.String())
		}
	}

	logger.Successf("Received %s", value.FormattedString())
	return nil
}
