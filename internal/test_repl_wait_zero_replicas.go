package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroReplicas(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()
	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run("--port", "6379"); err != nil {
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

	err = client.Wait("0", "60000", 0)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
