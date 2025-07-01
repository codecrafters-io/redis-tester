package test_cases

import (
	"fmt"
	"strings"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
	"github.com/codecrafters-io/tester-utils/logger"
)

type ClientUniqueTestCase struct {
	*SendCommandTestCase
	Client       *resp_client.RespConnection
	ExpectResult bool
}

/* This is useful when you have multiple clients which issue a blocking command (for eg. SUBSCRIBE, LPOP),
and a single client that issues a command which should unblock one or multiple clients
*/

type BlockingCommandTestCase struct {
	BlockingClientsTestCases []ClientUniqueTestCase
	ReleasingClientTestCase  *ClientUniqueTestCase
}

func (t *BlockingCommandTestCase) Run(logger *logger.Logger) error {
	if err := t.sendBlockingCommands(); err != nil {
		return err
	}
	err := t.ReleasingClientTestCase.Run(t.ReleasingClientTestCase.Client, logger)
	if err != nil {
		return err
	}
	if err := t.handleExpectingClients(logger); err != nil {
		return err
	}
	return t.verifyIdleClients(logger)
}

func (t *BlockingCommandTestCase) sendBlockingCommands() error {
	// Send commands from blocking clients synchronously
	// We do this to maintain order in fixtures
	for _, tc := range t.BlockingClientsTestCases {
		command := strings.ToUpper(tc.Command)
		if err := tc.Client.SendCommand(command, tc.Args...); err != nil {
			return err
		}
	}
	return nil
}

func (t *BlockingCommandTestCase) handleExpectingClients(logger *logger.Logger) error {
	logger.Infof("Checking responses in the blocked clients")
	// we do this synchronously to maintain fixtures order
	expectingClientsTestCase := Filter(t.BlockingClientsTestCases, func(tc ClientUniqueTestCase) bool {
		return tc.ExpectResult
	})

	for _, tc := range expectingClientsTestCase {
		value, err := tc.Client.ReadValue()
		if err != nil {
			return err
		}
		if err = tc.ProcessResponse(value, tc.Client, logger); err != nil {
			return err
		}
	}
	return nil
}

func (t *BlockingCommandTestCase) verifyIdleClients(logger *logger.Logger) error {
	logger.Infof("Checking if unexpected clients receive a response from the server")
	// If any one of the unexpected clients receive a value, we fail the test case
	idleClientsTestCase := Filter(t.BlockingClientsTestCases, func(tc ClientUniqueTestCase) bool {
		return !tc.ExpectResult
	})

	errorChan := make(chan error, len(idleClientsTestCase))
	doneChan := make(chan bool, len(idleClientsTestCase))

	for _, tc := range idleClientsTestCase {
		go func(tc *ClientUniqueTestCase) {
			defer func() { doneChan <- true }()

			// we don't need locking here because as soon as any client receive a response, we return an error
			value, err := tc.Client.ReadValue()
			if err == nil {
				// This is helpful in cases where the users might fan-out the response to all clients.
				// Printing the RESP value is more informative than printing the buffer
				errorChan <- fmt.Errorf("%s unexpectedly received value: %s",
					tc.Client.GetIdentifier(), value.FormattedString())
				return
			}
			if resp_decoder.IsEmptyContentReceivedError(err) {
				return
			}
			if tc.Client.UnreadBuffer.Len() > 0 {
				// decoding error
				errorChan <- fmt.Errorf("%s unexpectedly received: %q",
					tc.Client.GetIdentifier(), tc.Client.UnreadBuffer.String())
			} else {
				errorChan <- fmt.Errorf("received error while trying to receive a response for %s: %s",
					tc.Client.GetIdentifier(), err)
			}
		}(&tc)
	}

	completed := 0
	for completed < len(idleClientsTestCase) {
		select {
		case err := <-errorChan:
			return err
		case <-doneChan:
			completed++
		}
	}

	return nil
}
