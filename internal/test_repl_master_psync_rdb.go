package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplMasterPsyncRdb(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
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

	err = replica.Handshake()
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}
