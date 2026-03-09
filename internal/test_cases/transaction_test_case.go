package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// TransactionTestCase is a test case where we initiate a transaction by sending "MULTI" command
// Send a series of commands to the server expected back "QUEUED" for each command
// Finally send "EXEC" command and expect the response to be the same as ExpectedResponseArray
//
// RunAll will run all the steps in the Transaction execution. Alternatively, you
// can run each step individually.
type TransactionTestCase struct {
	// All the CommandQueue will be sent in order to client
	// And a string "QUEUED" will be expected
	CommandQueue [][]string

	// After queueing all the commands,
	// if "EXEC" is sent (based on which function is called)
	// The elements in the response array are asserted based on the
	// assertions in the  ExpectedResponseArray
	ExpectedResponseArray []resp_assertions.RESPAssertion
}

func (t TransactionTestCase) RunAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	if err := t.RunMulti(client, logger); err != nil {
		return err
	}

	if err := t.RunQueueAll(client, logger); err != nil {
		return err
	}

	if err := t.RunExec(client, logger); err != nil {
		return err
	}

	return nil
}

func (t TransactionTestCase) RunWithoutExec(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	if err := t.RunMulti(client, logger); err != nil {
		return err
	}

	if err := t.RunQueueAll(client, logger); err != nil {
		return err
	}

	return nil
}

func (t TransactionTestCase) RunMulti(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	commandTest := SendCommandTestCase{
		Command:   "MULTI",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	return commandTest.Run(client, logger)
}

func (t TransactionTestCase) RunQueueAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	for i, v := range t.CommandQueue {
		logger.Debugf("Sending command: %d/%d", i+1, len(t.CommandQueue))
		commandTest := SendCommandTestCase{
			Command:   v[0],
			Args:      v[1:],
			Assertion: resp_assertions.NewSimpleStringAssertion("QUEUED"),
		}
		if err := commandTest.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}

func (t TransactionTestCase) RunExec(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	var assertion resp_assertions.RESPAssertion

	// Expect a nil array if t.ExpectedResponseArray is nil
	// TODO: I'll remove this comment after PR review
	// The reasoning here was that Go allows us to distinguish between empty array / nil array (just like redis distinguishes these two)
	// using its own language construct (empty slice / nil slice)
	// So, I decided to implement it this way
	if t.ExpectedResponseArray == nil {
		assertion = resp_assertions.NewNilArrayAssertion()
	} else {
		assertion = resp_assertions.NewOrderedArrayAssertion(t.ExpectedResponseArray)
	}

	setCommandTestCase := SendCommandTestCase{
		Command:   "EXEC",
		Args:      []string{},
		Assertion: assertion,
	}

	return setCommandTestCase.Run(client, logger)
}
