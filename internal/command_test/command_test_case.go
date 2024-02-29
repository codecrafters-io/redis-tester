package command_test

import (
	"fmt"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/client"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

// TODO: Rename to SendCommandTestCase
type CommandTestCase struct {
	Command                   string
	Args                      []string
	Assertion                 resp_assertions.RESPAssertion
	ShouldSkipUnreadDataCheck bool
}

func (t CommandTestCase) Run(client *resp_client.RespClient, logger *logger.Logger) error {
	if err := client.SendCommand(t.Command, t.Args...); err != nil {
		return err
	}

	value, err := client.ReadValue()
	if err != nil {
		return err
	}

	if err = t.Assertion.Run(value); err != nil {
		return err
	}

	if !t.ShouldSkipUnreadDataCheck {
		client.ReadIntoBuffer() // Let's make sure there's no extra data

		if client.UnreadBuffer.Len() > 0 {
			return fmt.Errorf("Found extra data: %q", string(client.LastValueBytes)+client.UnreadBuffer.String())
		}
	}

	logger.Successf("Received %s", value.FormattedString())
	return nil
}
