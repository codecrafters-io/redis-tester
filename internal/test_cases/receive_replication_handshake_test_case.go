package test_cases

import (
	"encoding/hex"
	"fmt"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_encoder "github.com/codecrafters-io/redis-tester/internal/resp/encoder"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

// ReceiveReplicationHandshakeTestCase is a test case where we connect to a master
// as a replica and perform either all or a subset of the replication handshake.
//
// RunAll will run all the steps in the replication handshake. Alternatively, you
// can run each step individually.
type ReceiveReplicationHandshakeTestCase struct{}

func (t ReceiveReplicationHandshakeTestCase) RunAll(client *resp_connection.RespConnection, logger *logger.Logger) error {
	if err := t.RunPingStep(client, logger); err != nil {
		return err
	}

	if err := t.RunReplconfStep1(client, logger); err != nil {
		return err
	}

	if err := t.RunReplconfStep2(client, logger); err != nil {
		return err
	}

	if err := t.RunPsyncStep(client, logger); err != nil {
		return err
	}

	if err := t.RunSendRDBStep(client, logger); err != nil {
		return err
	}

	return nil
}

func (t ReceiveReplicationHandshakeTestCase) RunPingStep(client *resp_connection.RespConnection, logger *logger.Logger) error {
	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewCommandAssertion("PING"),
		Response:  resp_value.NewSimpleStringValue("PONG"),
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunReplconfStep1(client *resp_connection.RespConnection, logger *logger.Logger) error {
	commandTest := ReceiveCommandTestCase{
		Assertion:                 resp_assertions.NewCommandAssertion("REPLCONF", "listening-port", "6380"),
		Response:                  resp_value.NewSimpleStringValue("OK"),
		ShouldSkipUnreadDataCheck: true,
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunReplconfStep2(client *resp_connection.RespConnection, logger *logger.Logger) error {
	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewWildcardCommandAssertion("REPLCONF", "capa", "*", "?capa", "*"),
		Response:  resp_value.NewSimpleStringValue("OK"),
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunPsyncStep(client *resp_connection.RespConnection, logger *logger.Logger) error {
	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewCommandAssertion("PSYNC", "?", "-1"),
		Response:  resp_value.NewSimpleStringValue("FULLRESYNC IDIDIDIDIDIDIDIDIDIDIDIDIDIDIDIDIDIDIDID 0"),
	} // ToDo Add random ID generation

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunSendRDBStep(client *resp_connection.RespConnection, logger *logger.Logger) error {
	logger.Debugln("Sending RDB file...")

	hexStr := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Printf("Encountered %s while deconding hex string", err.Error())
		error := resp_value.NewErrorValue(err.Error())
		client.SendBytes(error.Bytes())
		return err
	}

	encodedValue := resp_encoder.Encode(resp_value.NewRDBAsBulkStringValue(string(bytes)))
	client.SendBytes(encodedValue)

	logger.Successf("Sent RDB file.")
	return nil
}
