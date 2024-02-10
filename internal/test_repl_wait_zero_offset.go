package internal

import (
	"fmt"
	"strconv"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testWaitZeroOffset(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	replicaCount := testerutils_random.RandomInt(3, 9)
	// replicas can be : [3, 9]
	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	for i := 0; i < replicaCount; i++ {
		logger.Debugf("Creating replica : %v.", i)
		conn, err := NewRedisConn("", "localhost:6379")
		defer conn.Close()

		if err != nil {
			fmt.Println("Error connecting to TCP server:", err)
		}

		replica := NewFakeRedisReplica(conn, logger)
		replica.LogPrefix = fmt.Sprintf("[Replica: %v] ", i+1)
		err = replica.Handshake()
		if err != nil {
			return err
		}
	}

	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[Client] "

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
