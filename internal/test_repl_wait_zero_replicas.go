package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testWaitZeroReplicas(stageHarness *testerutils.StageHarness) error {
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
	}

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[Client] "

	err = client.Wait("0", "60000", 0)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
