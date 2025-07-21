package test_cases

import (
	"bytes"
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
	rdb_parser "github.com/hdt3213/rdb/parser"
)

// SendReplicationHandshakeTestCase is a test case where we connect to a master
// as a replica and perform either all or a subset of the replication handshake.
//
// RunAll will run all the steps in the replication handshake. Alternatively, you
// can run each step individually.
type SendReplicationHandshakeTestCase struct{}

func (t SendReplicationHandshakeTestCase) RunAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger, listeningPort int) error {
	if err := t.RunPingStep(client, logger); err != nil {
		return err
	}

	if err := t.RunReplconfStep(client, logger, listeningPort); err != nil {
		return err
	}

	if err := t.RunPsyncStep(client, logger); err != nil {
		return err
	}

	if err := t.RunReceiveRDBStep(client, logger); err != nil {
		return err
	}

	return nil
}

func (t SendReplicationHandshakeTestCase) RunPingStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	commandTest := SendCommandTestCase{
		Command:   "PING",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("PONG"),
	}

	return commandTest.Run(client, logger)
}

func (t SendReplicationHandshakeTestCase) RunReplconfStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger, listeningPort int) error {
	commandTest := SendCommandTestCase{
		Command:   "REPLCONF",
		Args:      []string{"listening-port", fmt.Sprintf("%d", listeningPort)},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	if err := commandTest.Run(client, logger); err != nil {
		return err
	}

	commandTest = SendCommandTestCase{
		Command:   "REPLCONF",
		Args:      []string{"capa", "psync2"},
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}

	return commandTest.Run(client, logger)
}

func (t SendReplicationHandshakeTestCase) RunPsyncStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	commandTest := SendCommandTestCase{
		Command:                   "PSYNC",
		Args:                      []string{"?", "-1"},
		Assertion:                 resp_assertions.NewRegexStringAssertion("FULLRESYNC \\w+ 0"),
		ShouldSkipUnreadDataCheck: true, // We're expecting the RDB file to be sent next
	}

	return commandTest.Run(client, logger)
}

func (t SendReplicationHandshakeTestCase) RunReceiveRDBStep(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	clientLogger := client.GetLogger()
	clientLogger.Debugln("Reading RDB file...")

	rdbFileBytes, err := client.ReadFullResyncRDBFile()
	if err != nil {
		return err
	}

	// We don't care about the contents of the RDB file, we just want to make sure the file was valid
	processRedisObject := func(_ rdb_parser.RedisObject) bool {
		return true
	}

	decoder := rdb_parser.NewDecoder(bytes.NewReader(rdbFileBytes))
	if err = decoder.Parse(processRedisObject); err != nil {
		return fmt.Errorf("Invalid RDB file: %v", err)
	}

	client.ReadIntoBuffer() // Let's make sure there's no extra data

	if client.UnreadBuffer.Len() > 0 {
		return fmt.Errorf("Found extra data: %q", client.UnreadBuffer.String())
	}

	clientLogger.Successf("Received RDB file")
	return nil
}
