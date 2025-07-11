package test_cases

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
)

type NoResponseTestCase struct{}

func (n *NoResponseTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection) error {
	client.ReadIntoBuffer()
	if client.UnreadBuffer.Len() > 0 {
		return fmt.Errorf("%s received unexpected response: %q", client.GetIdentifier(), client.UnreadBuffer.String())
	}
	return nil
}
