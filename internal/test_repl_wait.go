package internal

import (
	"fmt"
	"strconv"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	logger "github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

// In this stage, we:
// 1. Boot the user's code as a Redis master.
// 2. Spawn multiple replicas and have each perform a handshake with the master.
// 3. ???
func testWait(stageHarness *testerutils.StageHarness) error {
	logger := stageHarness.Logger

	// Step 1: Boot the user's code as a Redis master.
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	// Step 2: Spawn multiple replicas and have each perform a handshake
	replicaCount := testerutils_random.RandomInt(3, 5)
	logger.Infof("Creating %v replicas.", replicaCount)

	replicas, err := spawnReplicas(replicaCount, logger)
	if err != nil {
		return err
	}

	// Step 3: Connect to master and issue a write command

	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
		return err
	}
	defer conn.Close()

	// TODO: Naming issue - this isn't a master?
	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[client] "

	client.SendAndAssert([]string{"SET", "foo", "123"}, []string{"OK"})

	// Step 4: Issue a WAIT command with 1 as the expected number of replicas
	err = client.Send([]string{"WAIT", "1", "500"})
	if err != nil {
		return err
	}

	masterOffset := 0

	// Step 5: Read propagated command + GETACK on all replicas
	for i := 0; i < replicaCount; i++ {
		replica := replicas[i]

		// Redis will send SELECT, but not expected from Users.
		err, _ = replica.readAndAssertMessagesWithSkip([]string{"SET", "foo", "123"}, "SELECT", true)
		if err != nil {
			return err
		}

		err, offsetDelta := replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, true)
		masterOffset += offsetDelta
		if err != nil {
			return err
		}
	}

	// Step 6: Issue ACKs from a subset of replicas
	replicaAcksCount := 1

	for i := 0; i < replicaAcksCount; i++ {
		replica := replicas[i]
		replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(masterOffset)})
	}

	// Step 7: Read the response of the WAIT command, assert it matches replicaAcksCount
	err = client.readAndAssertIntMessage(replicaAcksCount)
	if err != nil {
		return err
	}

	// Step 8: Issue another write command
	client.SendAndAssert([]string{"SET", "baz", "789"}, []string{"OK"})

	replicaAcksCount = min(replicaCount, testerutils_random.RandomInt(2, 6))
	timeout := 2000
	sendCount := strconv.Itoa(replicaAcksCount + 1)
	startTimeMilli := time.Now().UnixMilli()
	err = client.Send([]string{"WAIT", sendCount, strconv.Itoa(timeout)})
	if err != nil {
		return err
	}

	previousMasterOffset := masterOffset

	// Step 9: Read propagated command + GETACK on all replicas
	for i := 0; i < replicaCount; i++ {
		masterOffset = previousMasterOffset
		replica := replicas[i]

		err, o := replica.readAndAssertMessages([]string{"SET", "baz", "789"}, true)
		masterOffset += o

		err, o = replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, false)
		masterOffset += o
		if err != nil {
			return err
		}

		if i < replicaAcksCount {
			replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(masterOffset)})
		}
	}

	err = client.readAndAssertIntMessage(replicaAcksCount)
	if err != nil {
		return err
	}

	endTimeMilli := time.Now().UnixMilli()

	DELTA := 500 // ms
	timeElapsed := endTimeMilli - startTimeMilli
	client.Log(fmt.Sprintf("WAIT command returned after %v ms", timeElapsed))
	if timeElapsed > int64(timeout)+int64(DELTA) || timeElapsed < int64(timeout)-int64(DELTA) {
		return fmt.Errorf("Expected WAIT to return only after %v ms timeout elapsed.", timeout)
	}
	return nil
}

func spawnReplicas(replicaCount int, logger *logger.Logger) ([]*FakeRedisReplica, error) {
	var replicas []*FakeRedisReplica

	for i := 0; i < replicaCount; i++ {
		logger.Debugf("Creating replica #%v", i+1)

		conn, err := NewRedisConn("", "localhost:6379")
		if err != nil {
			fmt.Println("Error connecting to TCP server:", err)
			return nil, err
		}
		defer conn.Close()

		replica := NewFakeRedisReplica(conn, nil)
		replica.LogPrefix = fmt.Sprintf("[replica-%v] ", i+1)
		replicas = append(replicas, replica)
	}

	return replicas, nil
}
