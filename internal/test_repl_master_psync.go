package internal

import (
	"fmt"

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
	err = readAndAssertMessage(r, "PONG", logger)
	if err != nil {
		return err
	}

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")
	w.WriteCommand("REPLCONF", "listening-port", "6380")
	err = readAndAssertMessage(r, "OK", logger)
	if err != nil {
		return err
	}

	logger.Infof("$ redis-cli PSYNC ? -1")
	w.WriteCommand("PSYNC", "?", "-1")
	err = readAndAssertMessage(r, "FULLRESYNC * 0", logger)
	if err != nil {
		return err
	}

	return nil
}
