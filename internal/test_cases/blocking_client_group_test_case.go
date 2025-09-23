package test_cases

import (
	"container/list"
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type BlockingClientGroupTestCase struct {
	CommandToSend                 []string
	AssertionForReceivedResponse  resp_assertions.RESPAssertion
	ResponseExpectingClientsCount int
	Clients                       []*instrumented_resp_connection.InstrumentedRespConnection
}

func (t *BlockingClientGroupTestCase) SendBlockingCommands() error {
	command := t.CommandToSend[0]
	args := t.CommandToSend[1:]

	for _, client := range t.Clients {
		if err := client.SendCommand(command, args...); err != nil {
			return err
		}
	}

	// Ensure that the server processes all the Blocking commands first (so it can register all the waiting clients first)
	// and then only processes the next set of commands
	// Ordering does not matter; so we can sleep after all the commands have been sent
	time.Sleep(time.Millisecond)
	return nil
}

func (t *BlockingClientGroupTestCase) AssertResponses(logger *logger.Logger) error {
	// ensure that exactly 'ResponseExpectingClientsCount' clients receive the response
	// if the count is less -> error, more -> error
	clientWord := "client"
	if t.ResponseExpectingClientsCount != 1 {
		clientWord = "clients"
	}
	logger.Infof("Expecting %d %s to receive response of %s command", t.ResponseExpectingClientsCount, clientWord, t.CommandToSend[0])

	receivedResponsesCount := 0

	// Use a doubly linked list for easy removal during iteration
	clientsToProcess := list.New()
	for _, client := range t.Clients {
		clientsToProcess.PushFront(client)
	}

	errorChannel := make(chan error)
	allReadChannel := make(chan bool)

	go t.expectClientResponses(clientsToProcess, receivedResponsesCount, errorChannel, allReadChannel, logger)

	// Wait for allRead/error/timeout
	select {
	// In case of error, return error
	case err := <-errorChannel:
		return err
	// Return error if no response on any clients
	// The deadline of 2 second is choosen because client.ReadValue() has the default timeout of 2 seconds
	// It should be enough to infer for practical reasons that no response will be received now.
	case <-time.After(2 * time.Second):
		return fmt.Errorf("No response received in clients after a timeout of 2 seconds")
	// If response has been received from all expected channels, move to last check
	case <-allReadChannel:
		break
	}

	// Sleep for a small duration so if a wrong client receives a response even a little while later
	// we don't miss it
	time.Sleep(1 * time.Millisecond)

	// Check if any clients receive extra response
	for element := clientsToProcess.Front(); element != nil; element = element.Next() {
		client := element.Value.(*instrumented_resp_connection.InstrumentedRespConnection)
		receiveValueTestCase := NoResponseTestCase{}

		if err := receiveValueTestCase.Run(client); err != nil {
			return err
		}
	}

	return nil
}

func (t *BlockingClientGroupTestCase) expectClientResponses(
	clientsToProcess *list.List,
	receivedResponsesCount int,
	errorChannel chan error,
	allReadChannel chan bool,
	logger *logger.Logger,
) {
	// Loop until expected number of clients have received a valid response
	for receivedResponsesCount < t.ResponseExpectingClientsCount {

		// Iterate through clients in the list
		for element := clientsToProcess.Front(); element != nil; {
			client := element.Value.(*instrumented_resp_connection.InstrumentedRespConnection)
			nextClient := element.Next() // Store next element before potential removal

			// Check if client has received a response

			// Ignore errors because it's okay client may not have received any response yet
			// They may receive it in the future
			client.ReadIntoBuffer()

			// If no response, move to next client
			if client.UnreadBuffer.Len() == 0 {
				element = nextClient
				continue
			}

			// If response received, run the assertion against the received response
			receiveValueTestCase := ReceiveValueTestCase{
				Assertion: t.AssertionForReceivedResponse,
			}

			if err := receiveValueTestCase.Run(client, logger); err != nil {
				errorChannel <- err
				return
			}

			// Remove this client from the processing list
			clientsToProcess.Remove(element)

			receivedResponsesCount += 1
			// If correct response has been received from expected number of clients,
			// notify the main routine and return
			if receivedResponsesCount == t.ResponseExpectingClientsCount {
				allReadChannel <- true
				return
			}

			element = nextClient
		}
	}
}
