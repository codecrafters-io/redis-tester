package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
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
		return err
	}

	replica := NewFakeRedisReplica(conn, logger)

	err = replica.Ping()
	if err != nil {
		return err
	}

	err = replica.ReplConfPort()
	if err != nil {
		return err
	}

	err = replica.Psync()
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}
