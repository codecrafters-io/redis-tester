package test_cases

import (
	"fmt"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/tester-utils/logger"
)

type NoResponseTestCase struct{}

func (n *NoResponseTestCase) Run(client *resp_connection.RespConnection, logger *logger.Logger) error {
	client.ReadIntoBuffer()
	if client.UnreadBuffer.Len() > 0 {
		return fmt.Errorf("%s received unexpected response: %q", client.GetIdentifier(), client.UnreadBuffer.String())
	}
	return nil
}
