package test_cases

import (
	"fmt"
	"strings"

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

	result := t.Assertion.Run(value)
	if result.IsFailure() {
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
