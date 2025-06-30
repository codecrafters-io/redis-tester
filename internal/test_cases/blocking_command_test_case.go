package test_cases

import (
	"fmt"
	"sync"
	"time"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// MultiClientBlockingCommandTestCase represents a test case where multiple clients send blocking commands,
// one unlocker client sends a command to unblock them, and only specific clients should respond
type MultiClientBlockingCommandTestCase struct {
	// Blocking command details
	BlockingCommand   string
	BlockingArgs      []string
	BlockingAssertion resp_assertions.RESPAssertion

	// Unlocker command details
	UnlockerCommand   string
	UnlockerArgs      []string
	UnlockerAssertion resp_assertions.RESPAssertion

	// Client configuration
	BlockingClients []*resp_client.RespConnection
	UnlockerClient  *resp_client.RespConnection

	// Only clients at these indexes should respond to the blocking command
	ResponderIndexes []int

	// Timeout configuration
	TimeoutDuration *time.Duration

	// Internal state
	results     []error
	resultMutex sync.Mutex
	started     bool
}

// NewMultiClientBlockingCommandTestCase creates a new multi-client blocking command test case
func NewMultiClientBlockingCommandTestCase(
	blockingCommand string,
	blockingArgs []string,
	blockingAssertion resp_assertions.RESPAssertion,
	unlockerCommand string,
	unlockerArgs []string,
	unlockerAssertion resp_assertions.RESPAssertion,
	blockingClients []*resp_client.RespConnection,
	unlockerClient *resp_client.RespConnection,
	responderIndexes []int,
	timeout *time.Duration,
) *MultiClientBlockingCommandTestCase {
	var t time.Duration
	if timeout == nil {
		t = time.Second * 10
	} else {
		t = *timeout
	}

	return &MultiClientBlockingCommandTestCase{
		BlockingCommand:   blockingCommand,
		BlockingArgs:      blockingArgs,
		BlockingAssertion: blockingAssertion,
		UnlockerCommand:   unlockerCommand,
		UnlockerArgs:      unlockerArgs,
		UnlockerAssertion: unlockerAssertion,
		BlockingClients:   blockingClients,
		UnlockerClient:    unlockerClient,
		ResponderIndexes:  responderIndexes,
		TimeoutDuration:   &t,
		results:           make([]error, len(blockingClients)),
	}
}

// GetTimeout returns the timeout duration for the test case
func (t *MultiClientBlockingCommandTestCase) GetTimeout() time.Duration {
	if t.TimeoutDuration == nil {
		return time.Second * 10
	}
	return *t.TimeoutDuration
}

// Run starts all blocking commands on their respective clients
func (t *MultiClientBlockingCommandTestCase) Run(logger *logger.Logger) {
	if t.started {
		panic("MultiClientBlockingCommandTestCase already started")
	}
	t.started = true

	logger.Infof("Starting blocking commands on %d clients", len(t.BlockingClients))

	// Start blocking commands on all clients
	for i, client := range t.BlockingClients {
		go func(clientIndex int, client *resp_client.RespConnection) {
			blockingTestCase := SendCommandTestCase{
				Command:                   t.BlockingCommand,
				Args:                      t.BlockingArgs,
				Assertion:                 t.BlockingAssertion,
				ShouldSkipUnreadDataCheck: true,
			}

			err := blockingTestCase.Run(client, logger)

			t.resultMutex.Lock()
			t.results[clientIndex] = err
			t.resultMutex.Unlock()
		}(i, client)
	}
}

// SendUnlockerCommand sends the unlocker command and waits for its response
func (t *MultiClientBlockingCommandTestCase) SendUnlockerCommand(logger *logger.Logger) error {
	if !t.started {
		return fmt.Errorf("must call Run() before SendUnlockerCommand()")
	}

	logger.Infof("Sending unlocker command: %s %v", t.UnlockerCommand, t.UnlockerArgs)

	unlockerTestCase := SendCommandTestCase{
		Command:   t.UnlockerCommand,
		Args:      t.UnlockerArgs,
		Assertion: t.UnlockerAssertion,
	}

	return unlockerTestCase.Run(t.UnlockerClient, logger)
}

// WaitForResults waits for all blocking commands to complete and returns results
func (t *MultiClientBlockingCommandTestCase) WaitForResults() ([]error, error) {
	if !t.started {
		return nil, fmt.Errorf("must call Run() before WaitForResults()")
	}

	timeoutChan := time.After(t.GetTimeout())

	// Wait for all results or timeout
	for {
		t.resultMutex.Lock()
		allComplete := true
		for _, result := range t.results {
			if result == nil {
				allComplete = false
				break
			}
		}
		t.resultMutex.Unlock()

		if allComplete {
			break
		}

		select {
		case <-timeoutChan:
			return t.results, fmt.Errorf("timeout waiting for blocking command results")
		case <-time.After(10 * time.Millisecond):
			// Continue checking
		}
	}

	return t.results, nil
}

// ValidateResults checks that only the expected clients responded and others didn't
func (t *MultiClientBlockingCommandTestCase) ValidateResults(logger *logger.Logger) error {
	results, err := t.WaitForResults()
	if err != nil {
		return err
	}

	// Create a set of responder indexes for quick lookup
	responderSet := make(map[int]bool)
	for _, idx := range t.ResponderIndexes {
		responderSet[idx] = true
	}

	// Check that only expected clients responded
	for i, result := range results {
		shouldRespond := responderSet[i]
		didRespond := result == nil // nil means success

		if shouldRespond && !didRespond {
			return fmt.Errorf("client %d should have responded but failed: %v", i, result)
		}

		if !shouldRespond && didRespond {
			return fmt.Errorf("client %d should not have responded but succeeded", i)
		}

		if shouldRespond {
			logger.Successf("Client %d responded as expected", i)
		} else {
			logger.Successf("Client %d correctly did not respond", i)
		}
	}

	return nil
}

// RunComplete executes the entire test case: start blocking commands, send unlocker command, and validate results
func (t *MultiClientBlockingCommandTestCase) RunComplete(logger *logger.Logger) error {
	t.Run(logger)

	// Give blocking commands a moment to start
	time.Sleep(100 * time.Millisecond)

	if err := t.SendUnlockerCommand(logger); err != nil {
		return fmt.Errorf("unlocker command failed: %v", err)
	}

	return t.ValidateResults(logger)
}

// Legacy BlockingCommandTestCase for backward compatibility
type BlockingCommandTestCase struct {
	SendCommandTestCase
	timeoutDuration *time.Duration
	resumeChannel   chan struct{}
	resultChan      chan error
}

func NewBlockingCommandTestCase(command string, args []string, assertion resp_assertions.RESPAssertion, timeout *time.Duration) *BlockingCommandTestCase {
	var t time.Duration
	if timeout == nil {
		t = time.Second * 10
	} else {
		t = *timeout
	}
	return &BlockingCommandTestCase{
		SendCommandTestCase: SendCommandTestCase{
			Command:                   command,
			Args:                      args,
			Assertion:                 assertion,
			ShouldSkipUnreadDataCheck: true,
		},
		resumeChannel:   make(chan struct{}, 1),
		resultChan:      make(chan error, 1),
		timeoutDuration: &t,
	}
}

func (t *BlockingCommandTestCase) GetTimeout() time.Duration {
	if t.timeoutDuration == nil {
		return time.Second * 10
	}
	return *t.timeoutDuration
}

func (t *BlockingCommandTestCase) Run(client *resp_client.RespConnection, logger *logger.Logger) {
	go func() {
		responseChannel := make(chan error)
		timeoutChan := time.After(t.GetTimeout())
		go func() {
			err := t.SendCommandTestCase.Run(client, logger)
			responseChannel <- err
		}()

		select {
		case <-t.resumeChannel:
			t.resultChan <- <-responseChannel
		case <-timeoutChan:
			t.resultChan <- <-responseChannel
		}
	}()
}

func (t *BlockingCommandTestCase) WaitForResult() error {
	return <-t.resultChan
}

func (t *BlockingCommandTestCase) Resume() {
	select {
	case t.resumeChannel <- struct{}{}:
	default:
		panic("Codecrafters Internal Error - blocking test case already resumed")
	}
}
