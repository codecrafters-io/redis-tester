package test_cases

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_encoder "github.com/codecrafters-io/redis-tester/internal/resp/encoder"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// ReceiveReplicationHandshakeTestCase is a test case where we connect to a master
// as a replica and perform either all or a subset of the replication handshake.
//
// RunAll will run all the steps in the replication handshake. Alternatively, you
// can run each step individually.
type ReceiveReplicationHandshakeTestCase struct{}

func (t ReceiveReplicationHandshakeTestCase) RunAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
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

func (t ReceiveReplicationHandshakeTestCase) RunPingStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	client.GetLogger().Infof("Waiting for replica to initiate handshake with %q command", "PING")

	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewCommandAssertion("PING"),
		Response:  resp_value.NewSimpleStringValue("PONG"),
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunReplconfStep1(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	client.GetLogger().Infof("Waiting for replica to send %q command", "REPLCONF listening-port 6380")

	commandTest := ReceiveCommandTestCase{
		Assertion:                 resp_assertions.NewCommandAssertion("REPLCONF", "listening-port", "6380"),
		Response:                  resp_value.NewSimpleStringValue("OK"),
		ShouldSkipUnreadDataCheck: true,
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunReplconfStep2(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	client.GetLogger().Infof("Waiting for replica to send %q command", "REPLCONF capa")

	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewOnlyCommandAssertion("REPLCONF"),
		Response:  resp_value.NewSimpleStringValue("OK"),
	}

	err := commandTest.Run(client, logger)
	if err != nil {
		return err
	}

	receivedValue := commandTest.ReceivedValue

	elements := receivedValue.Array()

	if len(elements) < 3 {
		return fmt.Errorf("Expected array with at least 3 element, got %d elements", len(elements))
	}

	firstCapaArg := elements[1].String()

	if elements[1].Type != resp_value.SIMPLE_STRING && elements[1].Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected first replconf argument to be a string, got %s", elements[1].Type)
	}

	if !strings.EqualFold(firstCapaArg, "capa") {
		return fmt.Errorf("Expected first replconf argument to be %q, got %q", "capa", strings.ToLower(firstCapaArg))
	}

	if len(elements) == 5 {
		if elements[3].Type != resp_value.SIMPLE_STRING && elements[3].Type != resp_value.BULK_STRING {
			return fmt.Errorf("Expected third replconf argument to be a string, got %s", elements[3].Type)
		}

		secondCapaArg := elements[3].String()

		if !strings.EqualFold(secondCapaArg, "capa") {
			return fmt.Errorf("Expected third replconf argument to be %q, got %q", "capa", strings.ToLower(secondCapaArg))
		}
	}

	return nil
}

func (t ReceiveReplicationHandshakeTestCase) RunPsyncStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	client.GetLogger().Infof("Waiting for replica to send %q command", "PSYNC")

	id := "75cd7bc10c49047e0d163660f3b90625b1af31dc"
	commandTest := ReceiveCommandTestCase{
		Assertion: resp_assertions.NewCommandAssertion("PSYNC", "?", "-1"),
		Response:  resp_value.NewSimpleStringValue(fmt.Sprintf("FULLRESYNC %v 0", id)),
	}

	return commandTest.Run(client, logger)
}

func (t ReceiveReplicationHandshakeTestCase) RunSendRDBStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	clientLogger := client.GetLogger()
	clientLogger.Debugln("Sending RDB file...")

	hexStr := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(fmt.Sprintf("Encountered %s while decoding hex string", err.Error()))
	}

	encodedValue := resp_encoder.EncodeFullResyncRDBFile(bytes)
	client.SendBytes(encodedValue)

	clientLogger.Successf("Sent RDB file.")
	return nil
}
