package test_cases

import (
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
	receiveValueTestCase := ReceiveValueTestCase{
		Assertion:                 t.Assertion,
		ShouldSkipUnreadDataCheck: t.ShouldSkipUnreadDataCheck,
	}
	err := receiveValueTestCase.Run(conn, logger)
	if err != nil {
		return err
	}
	t.ReceivedValue = receiveValueTestCase.ActualValue

	if err := conn.SendValue(t.Response); err != nil {
		return err
	}

	return nil
}
