package test_cases

import (
	"sync"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type clientAssertionBinding struct {
	Client  *instrumented_resp_connection.InstrumentedRespConnection
	Command string
	Args    []string

	// If nil, we expect no response
	Assertion *resp_assertions.RESPAssertion
}

type BlockingClientGroupTestCase struct {
	clientsWithExpectedAssertions []clientAssertionBinding
}

func (t *BlockingClientGroupTestCase) AddClientWithExpectedResponse(client *instrumented_resp_connection.InstrumentedRespConnection, command string, args []string, assertion resp_assertions.RESPAssertion) *BlockingClientGroupTestCase {
	t.clientsWithExpectedAssertions = append(t.clientsWithExpectedAssertions, clientAssertionBinding{
		Client:    client,
		Command:   command,
		Args:      args,
		Assertion: &assertion,
	})

	return t
}

func (t *BlockingClientGroupTestCase) AddClientWithNoExpectedResponse(client *instrumented_resp_connection.InstrumentedRespConnection, command string, args []string) *BlockingClientGroupTestCase {
	t.clientsWithExpectedAssertions = append(t.clientsWithExpectedAssertions, clientAssertionBinding{
		Client:    client,
		Command:   command,
		Args:      args,
		Assertion: nil,
	})

	return t
}

func (t *BlockingClientGroupTestCase) SendBlockingCommands() error {
	for _, clientWithExpectedResponse := range t.clientsWithExpectedAssertions {
		if err := clientWithExpectedResponse.Client.SendCommand(clientWithExpectedResponse.Command, clientWithExpectedResponse.Args...); err != nil {
			return err
		}
		time.Sleep(1 * time.Millisecond) // Ensure server receives commands in order
	}

	return nil
}

func (t *BlockingClientGroupTestCase) AssertResponses(logger *logger.Logger) error {
	if len(t.clientsWithExpectedAssertions) == 0 {
		return nil
	}

	// First, log which clients expect responses
	// clients which do not expect responses don't need logging because it's automatically handled by NoResponseTestCase
	for _, clientWithExpectedResponse := range t.clientsWithExpectedAssertions {
		if clientWithExpectedResponse.Assertion != nil {
			clientWithExpectedResponse.Client.GetLogger().Infof("Expecting response of %s command", clientWithExpectedResponse.Command)
		}
	}

	// Use sync.WaitGroup to handle test cases in any order
	var waitGroup sync.WaitGroup
	errorChan := make(chan error, len(t.clientsWithExpectedAssertions))

	for _, clientWithExpectedResponse := range t.clientsWithExpectedAssertions {
		waitGroup.Add(1)
		go func(clientWithExpectedResponse clientAssertionBinding) {
			defer waitGroup.Done()

			if clientWithExpectedResponse.Assertion == nil {
				// No response expected
				testCase := NoResponseTestCase{}
				if err := testCase.Run(clientWithExpectedResponse.Client); err != nil {
					errorChan <- err
				}
				return
			}
			// Response expected
			testCase := ReceiveValueTestCase{
				Assertion: *clientWithExpectedResponse.Assertion,
			}

			if err := testCase.Run(clientWithExpectedResponse.Client, logger); err != nil {
				errorChan <- err
			}
		}(clientWithExpectedResponse)
	}

	// Wait for all goroutines to complete
	go func() {
		waitGroup.Wait()
		close(errorChan)
	}()

	// Check for any errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}
