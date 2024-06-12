package test_cases

import (
	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// TransactionTestCase is a test case where we initiate a transaction by sending "MULTI" command
// Send a series of commands to the server expected back "QUEUED" for each command
// Finally send "EXEC" command and expect the response to be the same as ResultArray
//
// RunAll will run all the steps in the Transaction execution. Alternatively, you
// can run each step individually.
type TransactionTestCase struct {
	// All the CommandQueue will be sent in order to client
	// And a string "QUEUED" will be expected
	CommandQueue [][]string

	// After queueing all the commands,
	// if ResultArray is not empty, "EXEC" is sent
	// And the response is compared with this ResultArray
	ResultArray []resp_value.Value
}

func (t TransactionTestCase) RunAll(client *resp_client.RespConnection, logger *logger.Logger) error {
	if err := t.RunMulti(client, logger); err != nil {
		return err
	}

	if err := t.RunQueueAll(client, logger); err != nil {
		return err
	}

	if len(t.ResultArray) > 0 {
		if err := t.RunExec(client, logger); err != nil {
			return err
		}
	}

	return nil
}

func (t TransactionTestCase) RunMulti(client *resp_client.RespConnection, logger *logger.Logger) error {
	commandTest := SendCommandTestCase{
		Command:   "MULTI",
		Args:      []string{},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	return commandTest.Run(client, logger)
}

func (t TransactionTestCase) RunQueueAll(client *resp_client.RespConnection, logger *logger.Logger) error {
	for i, v := range t.CommandQueue {
		logger.Debugf("Sent #%d command", i)
		commandTest := SendCommandTestCase{
			Command:   v[0],
			Args:      v[1:],
			Assertion: resp_assertions.NewStringAssertion("QUEUED"),
		}
		if err := commandTest.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}

func (t TransactionTestCase) RunExec(client *resp_client.RespConnection, logger *logger.Logger) error {
	setCommandTestCase := SendCommandTestCase{
		Command:   "EXEC",
		Args:      []string{},
		Assertion: resp_assertions.NewOrderedArrayAssertion(t.ResultArray),
	}

	return setCommandTestCase.Run(client, logger)
}
