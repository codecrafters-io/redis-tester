package test_cases

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
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

func (t *ReceiveValueTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	err := t.RunWithoutAssert(client)
	if err != nil {
		return err
	}
	return t.Assert(client, logger)
}

func (t *ReceiveValueTestCase) RunWithoutAssert(client *instrumented_resp_connection.InstrumentedRespConnection) error {
	value, err := client.ReadValue()
	if err != nil {
		return err
	}
	t.ActualValue = value
	return nil
}

func (t *ReceiveValueTestCase) Assert(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	if err := t.Assertion.Run(t.ActualValue); err != nil {
		return err
	}

	if !t.ShouldSkipUnreadDataCheck {
		client.ReadIntoBuffer() // Let's make sure there's no extra data

		if client.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", client.UnreadBuffer.String())
		}
	}

	client.GetLogger().Successf("✔︎ Received %s", t.ActualValue.FormattedString())
	return nil
}
