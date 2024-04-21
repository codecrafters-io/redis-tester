package test_cases

import (
	"fmt"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

type ReceiveValueTestCase struct {
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool

	// This is set after the test is run
	ActualValue resp_value.Value
}

func (t *ReceiveValueTestCase) Run(client *resp_client.RespConnection, logger *logger.Logger) error {
	value, err := client.ReadValue()
	if err != nil {
		return err
	}

	t.ActualValue = value

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
