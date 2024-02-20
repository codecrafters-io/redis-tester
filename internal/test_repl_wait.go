package internal

import (
	"fmt"
	"math"
	"strconv"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

// In this stage, we:
//  1. Boot the user's code as a Redis master.
//  2. Spawn multiple replicas and have each perform a handshake with the master.
//  3. Issue a write command, test WAIT 1 500
//     3.1. Issue a write command to the master
//     3.2. Issue a WAIT command with 1 as the expected number of replicas
//     3.3. Read propagated command on replicas + respond to subset of GETACKs
//     3.4. Assert response of WAIT command is 1
//  4. Issue another write command, test WAIT <REPLICA_COUNT+1> 2000
//     4.1. Issue a write command to the master
//     4.2. Issue a WAIT command with a subset as the expected number of replicas
//     4.3. Read propagated command on replicas + respond to subset of GETACKs
//     4.4. Assert response of WAIT command is acks count
//     4.5. Assert that the WAIT command returned after the timeout
func testWait(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()

	// Step 1: Boot the user's code as a Redis master.
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// Step 2: Spawn multiple replicas and have each perform a handshake
	replicaCount := testerutils_random.RandomInt(3, 5)
	offset := 0

	logger.Infof("Creating %v replicas.", replicaCount)

	replicas, err := spawnReplicas(replicaCount, logger)
	if err != nil {
		return err
	}

	// Step 3.1: Connect to master and issue a write command
	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
		return err
	}
	defer conn.Close()

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[client] "

	client.SendAndAssert([]string{"SET", "foo", "123"}, []string{"OK"})

	// Step 3.2: Issue a WAIT command with 1 as the expected number of replicas
	err = client.Send([]string{"WAIT", "1", "500"})
	if err != nil {
		return err
	}

	replicaAcksCount := 1
	masterOffset := offset
	// Step 3.3: Read propagated command on replicas + respond to subset of GETACKs
	for i := 0; i < replicaCount; i++ {
		offset = masterOffset
		replica := replicas[i]

		// Redis will send SELECT, but not expected from Users.
		err, o := replica.readAndAssertMessagesWithSkip([]string{"SET", "foo", "123"}, "SELECT", true)
		offset += o
		if err != nil {
			return err
		}

		err, o = replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, true)
		offset += o
		if err != nil {
			return err
		}

		if i < replicaAcksCount {
			replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
		}
	}

	// Step 3.4: Assert response of WAIT command is 1
	err = client.readAndAssertIntMessage(replicaAcksCount)
	if err != nil {
		return err
	}

	// Step 4.1: Issue another write command
	client.SendAndAssert([]string{"SET", "baz", "789"}, []string{"OK"})

	// Step 4.2: Issue a WAIT command with a subset as the expected number of replicas
	replicaAcksCount = testerutils_random.RandomInt(2, replicaCount)
	timeout := 2000
	replicaAcksWaitCount := strconv.Itoa(replicaAcksCount + 1)
	startTimeMilli := time.Now().UnixMilli()
	err = client.Send([]string{"WAIT", replicaAcksWaitCount, strconv.Itoa(timeout)})
	if err != nil {
		return err
	}

	masterOffset = offset
	// Step 4.3: Read propagated command on replicas + respond to subset of GETACKs
	for i := 0; i < replicaCount; i++ {
		offset = masterOffset
		replica := replicas[i]

		err, o := replica.readAndAssertMessages([]string{"SET", "baz", "789"}, true)
		offset += o

		err, o = replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, false)
		offset += o
		if err != nil {
			return err
		}

		if i < replicaAcksCount {
			replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
		}
	}

	// Step 4.4: Assert response of WAIT command is replicaAcksCount
	err = client.readAndAssertIntMessage(replicaAcksCount)
	if err != nil {
		return err
	}

	// Step 4.5: Assert that the WAIT command returned after the timeout
	endTimeMilli := time.Now().UnixMilli()

	threshold := 500 // ms
	timeElapsed := endTimeMilli - startTimeMilli
	client.Log(fmt.Sprintf("WAIT command returned after %v ms", timeElapsed))
	if math.Abs(float64(timeElapsed-int64(timeout))) > float64(threshold) {
		return fmt.Errorf("Expected WAIT to return only after %v ms timeout elapsed.", timeout)
	}
	return nil
}

func spawnReplicas(replicaCount int, logger *logger.Logger) ([]*FakeRedisReplica, error) {
	var replicas []*FakeRedisReplica

	for i := 0; i < replicaCount; i++ {
		logger.Debugf("Creating replica : %v.", i+1)
		conn, err := NewRedisConn("", "localhost:6379")
		if err != nil {
			fmt.Println("Error connecting to TCP server:", err)
			return nil, err
		}

		replica := NewFakeRedisReplica(conn, logger)
		replicas = append(replicas, replica)
		replica.LogPrefix = fmt.Sprintf("[replica-%v] ", i+1)

		err = replica.Handshake()
		if err != nil {
			return nil, err
		}
	}
	return replicas, nil
}
