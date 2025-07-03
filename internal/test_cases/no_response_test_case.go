package test_cases

import (
	"fmt"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
)

type NoResponseTestCase struct{}

func (n *NoResponseTestCase) Run(client *resp_connection.RespConnection) error {
	client.ReadIntoBuffer()
	if client.UnreadBuffer.Len() > 0 {
		return fmt.Errorf("%s received unexpected response: %q", client.GetIdentifier(), client.UnreadBuffer.String())
	}
	return nil
}
