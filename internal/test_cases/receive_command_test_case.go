package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
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

func (t *ReceiveCommandTestCase) Run(conn *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	receiveValueTestCase := ReceiveValueTestCase{
		Assertion:                 t.Assertion,
		ShouldSkipUnreadDataCheck: t.ShouldSkipUnreadDataCheck,
	}
	err := receiveValueTestCase.Run(conn, logger)
	t.ReceivedValue = receiveValueTestCase.ActualValue

	if err != nil {
		return err
	}
	if err := conn.SendValue(t.Response); err != nil {
		return err
	}

	return nil
}
