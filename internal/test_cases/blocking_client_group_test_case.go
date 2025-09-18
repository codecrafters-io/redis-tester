package test_cases

import (
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type clientWithExpectedResponse struct {
	Client  *instrumented_resp_connection.InstrumentedRespConnection
	Command string
	Args    []string

	// If nil, we expect no response
	Assertion *resp_assertions.RESPAssertion
}

type BlockingClientGroupTestCase struct {
	clientsWithExpectedResponses []clientWithExpectedResponse
}

func (t *BlockingClientGroupTestCase) AddClientWithExpectedResponse(client *instrumented_resp_connection.InstrumentedRespConnection, command string, args []string, assertion resp_assertions.RESPAssertion) *BlockingClientGroupTestCase {
	t.clientsWithExpectedResponses = append(t.clientsWithExpectedResponses, clientWithExpectedResponse{
		Client:    client,
		Command:   command,
		Args:      args,
		Assertion: &assertion,
	})

	return t
}

func (t *BlockingClientGroupTestCase) AddClientWithNoExpectedResponse(client *instrumented_resp_connection.InstrumentedRespConnection, command string, args []string) *BlockingClientGroupTestCase {
	t.clientsWithExpectedResponses = append(t.clientsWithExpectedResponses, clientWithExpectedResponse{
		Client:    client,
		Command:   command,
		Args:      args,
		Assertion: nil,
	})

	return t
}

func (t *BlockingClientGroupTestCase) SendBlockingCommands() error {
	for _, clientWithExpectedResponse := range t.clientsWithExpectedResponses {
		if err := clientWithExpectedResponse.Client.SendCommand(clientWithExpectedResponse.Command, clientWithExpectedResponse.Args...); err != nil {
			return err
		}
		time.Sleep(1 * time.Millisecond) // Ensure server receives commands in order
	}

	return nil
}

func (t *BlockingClientGroupTestCase) AssertResponses(logger *logger.Logger) error {
	for i := len(t.clientsWithExpectedResponses) - 1; i >= 0; i-- {
		clientWithExpectedResponse := t.clientsWithExpectedResponses[i]
		clientLogger := clientWithExpectedResponse.Client.GetLogger()
		if clientWithExpectedResponse.Assertion == nil {
			testCase := NoResponseTestCase{}
			if err := testCase.Run(clientWithExpectedResponse.Client); err != nil {
				return err
			}
		} else {
			clientLogger.Infof("Expecting response of %s command", clientWithExpectedResponse.Command)
			testCase := ReceiveValueTestCase{
				Assertion: *clientWithExpectedResponse.Assertion,
			}
			if err := testCase.Run(clientWithExpectedResponse.Client, logger); err != nil {
				return err
			}
		}
	}
	return nil
}
