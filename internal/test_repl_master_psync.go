package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplMasterPsync(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	r := resp3.NewReader(conn)
	w := resp3.NewWriter(conn)

	logger.Infof("$ redis-cli PING")

	w.WriteCommand("PING")
	actualMessage, err := readRespString(r, logger)
	if err != nil {
		return err
	}
	if actualMessage != "PONG" {
		return fmt.Errorf("Expected 'PONG', got %v", actualMessage)
	}
	logger.Successf("PONG received.")

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")
	w.WriteCommand("REPLCONF", "listening-port", "6380")
	actualMessage, err = readRespString(r, logger)
	if err != nil {
		return err
	}
	if actualMessage != "OK" {
		return fmt.Errorf("Expected 'OK', got %v", actualMessage)
	}
	logger.Successf("OK received.")

	w.WriteCommand("PSYNC", "?", "-1")
	actualMessage, err = readRespString(r, logger)
	actualMessageParts := strings.Split(actualMessage, " ")
	command, offset := actualMessageParts[0], actualMessageParts[2]
	if command != "FULLRESYNC" {
		return fmt.Errorf("Expected 'FULLRESYNC' from Master, got %v", command)
	}
	logger.Successf("FULLRESYNC received.")
	if offset != "0" {
		return fmt.Errorf("Expected Offset to be 0 from Master, got %v", offset)
	}
	logger.Successf("Offset = 0 received.")

	return nil
}
