package internal

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroReplicas(stageHarness *test_case_harness.TestCaseHarness) error {
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

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[client] "

	err = client.Wait(0, time.Millisecond*60000, 0)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
