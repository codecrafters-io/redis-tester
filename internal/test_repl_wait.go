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

	client := NewFakeRedisClient(conn, logger)
	client.LogPrefix = "[client] "

	client.SendAndAssert([]string{"SET", "foo", "123"}, []string{"OK"})

	// Step 3.2: Issue a WAIT command with 1 as the expected number of replicas
	err = client.Send([]string{"WAIT", "1", "500"})
	if err != nil {
		return err
	}

	masterOffset := 0
	replicaAcksCount := 1

	// Step 3.3: Read propagated command on replicas + respond to subset of GETACKs
	masterOffset, err = consumeReplicationStreamAndSendPartialAcks(replicas, replicaAcksCount, masterOffset, []string{"SET", "foo", "123"}, []string{"REPLCONF", "GETACK", "*"})
	if err != nil {
		return err
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
	waitCommandAcksCount := strconv.Itoa(replicaAcksCount + 1)
	startTimeMilli := time.Now().UnixMilli()
	timeout := 2000
	err = client.Send([]string{"WAIT", waitCommandAcksCount, strconv.Itoa(timeout)})
	if err != nil {
		return err
	}

	// Step 4.3: Read propagated command on replicas + respond to subset of GETACKs
	masterOffset, err = consumeReplicationStreamAndSendPartialAcks(replicas, replicaAcksCount, masterOffset, []string{"SET", "baz", "789"}, []string{"REPLCONF", "GETACK", "*"})
	if err != nil {
		return err
	}

	// Step 4.4: Assert response of WAIT command is replicaAcksCount
	err = client.readAndAssertIntMessage(replicaAcksCount)
	if err != nil {
		return err
	}

	// Step 4.5: Assert that the WAIT command returned after the timeout
	endTimeMilli := time.Now().UnixMilli()

	threshold := 500 // ms
	elapsedTimeMilli := endTimeMilli - startTimeMilli
	client.Log(fmt.Sprintf("WAIT command returned after %v ms", elapsedTimeMilli))
	if math.Abs(float64(elapsedTimeMilli-int64(timeout))) > float64(threshold) {
		return fmt.Errorf("Expected WAIT to return exactly after %v ms timeout elapsed.", timeout)
	}

	for _, replica := range replicas {
		replica.Conn.Close()
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

func consumeReplicationStreamAndSendPartialAcks(replicas []*FakeRedisReplica, replicaAcksCount int, previousMasterOffset int, firstCommand []string, secondCommand []string) (newMasterOffset int, err error) {
	for i := 0; i < len(replicas); i++ {
		replica := replicas[i]

		// Redis will send SELECT, but not expected from Users.
		// We skip the SELECT command, IF received.
		// Then check the next received command.
		err, offsetDeltaFromSetCommand := replica.readAndAssertMessagesWithSkip(firstCommand, "SELECT", true)
		if err != nil {
			return 0, err
		}

		err, offsetDeltaFromGetAckCommand := replica.readAndAssertMessages(secondCommand, false)
		if err != nil {
			return 0, err
		}

		newMasterOffset = previousMasterOffset + offsetDeltaFromSetCommand + offsetDeltaFromGetAckCommand

		if i < replicaAcksCount {
			replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(newMasterOffset)})
		}
	}

	return newMasterOffset, nil
}
