package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"strconv"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWaitZeroOffset(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()
	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run([]string{
		"--port", "6379",
	}...); err != nil {
		return err
	}

	logger := stageHarness.Logger

	replicaCount := testerutils_random.RandomInt(3, 9)
	// replicas can be : [3, 9]
	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	for i := 0; i < replicaCount; i++ {
		logger.Debugf("Creating replica : %v.", i)
		conn, err := NewRedisConn("", "localhost:6379")
		if err != nil {
			fmt.Println("Error connecting to TCP server:", err)
			return err
		}
		defer conn.Close()

		replica := NewFakeRedisReplica(conn, logger)
		replica.LogPrefix = fmt.Sprintf("[replica-%v] ", i+1)
		err = replica.Handshake()
		if err != nil {
			return err
		}
	}

	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
		return err
	}

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[client] "

	diff := ((replicaCount + 3) - 3) / 3
	safeDiff := max(1, diff) // If diff is 0, it will get stuck in an infinite loop
	for i := 3; i < replicaCount+3; i += safeDiff {
		actual, expected := strconv.Itoa(i), replicaCount
		err = client.Wait(actual, "500", expected)
		if err != nil {
			return err
		}
	}

	conn.Close()
	return nil
}
