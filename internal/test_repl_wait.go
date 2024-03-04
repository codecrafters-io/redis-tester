package internal

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type WaitTest struct {
	// WriteCommand is the command we'll issue to the master
	WriteCommand []string

	// WaitReplicaCount is the number of replicas we'll specify in the WAIT command
	WaitReplicaCount int

	// WaitTimeoutMilli is the timeout we'll specify in the WAIT command
	WaitTimeoutMilli int

	// ActualNumberOfAcks is the number of ACKs we'll send back to the master
	ActualNumberOfAcks int

	ShouldVerifyTimeout bool
}

// In this stage, we:
//  1. Boot the user's code as a Redis master.
//  2. Spawn multiple replicas and have each perform a handshake with the master.
//  3. Connect to Master, and execute RunWaitTest
//  4. RunWaitTest :
//     4.1. Issue a write command to the master
//     4.2. Issue a WAIT command with WaitReplicaCount as the expected number of replicas
//     4.3. Read propagated command on replicas + respond to subset of GETACKs
//     4.4. Assert response of WAIT command is ActualNumberOfAcks
//     4.5. Assert that the WAIT command returned after the timeout
func testWait(stageHarness *test_case_harness.TestCaseHarness) error {
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

	// Step 3: Connect to master
	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
		return err
	}
	defer conn.Close()

	client := NewFakeRedisClient(conn, logger)
	client.LogPrefix = "[client] "
	var masterOffset int

	replicaSubsetCount := 1
	if masterOffset, err = RunWaitTest(client, replicas, 0, WaitTest{
		WriteCommand:        []string{"SET", "foo", "123"},
		WaitReplicaCount:    replicaSubsetCount,
		ActualNumberOfAcks:  replicaSubsetCount,
		WaitTimeoutMilli:    500,
		ShouldVerifyTimeout: false,
	}); err != nil {
		return err
	}

	replicaSubsetCount = testerutils_random.RandomInt(2, replicaCount)
	if _, err = RunWaitTest(client, replicas, masterOffset, WaitTest{
		WriteCommand:        []string{"SET", "baz", "789"},
		WaitReplicaCount:    replicaSubsetCount + 1,
		ActualNumberOfAcks:  replicaSubsetCount,
		WaitTimeoutMilli:    2000,
		ShouldVerifyTimeout: true,
	}); err != nil {
		return err
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
		offsetDeltaFromSetCommand, err := replica.readAndAssertMessagesWithSkip(firstCommand, "SELECT", true)
		if err != nil {
			return 0, err
		}

		offsetDeltaFromGetAckCommand, err := replica.readAndAssertMessages(secondCommand, false)
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

func RunWaitTest(client *FakeRedisClient, replicas []*FakeRedisReplica, replicationOffset int, waitTest WaitTest) (newReplicationOffset int, err error) {
	// Step 1: Issue a write command
	client.SendAndAssertStringArray(waitTest.WriteCommand, []string{"OK"})

	// Step 2: Issue a WAIT command with a subset as the expected number of replicas
	startTimeMilli := time.Now().UnixMilli()
	err = client.Send([]string{"WAIT", strconv.Itoa(waitTest.WaitReplicaCount), strconv.Itoa(waitTest.WaitTimeoutMilli)})
	if err != nil {
		return 0, err
	}

	// Step 3: Read propagated command on replicas + respond to subset of GETACKs
	newReplicationOffset, err = consumeReplicationStreamAndSendPartialAcks(replicas, waitTest.ActualNumberOfAcks, replicationOffset, waitTest.WriteCommand, []string{"REPLCONF", "GETACK", "*"})
	if err != nil {
		return 0, err
	}

	// Step 4: Assert response of WAIT command is replicaAcksCount
	err = client.readAndAssertIntMessage(waitTest.ActualNumberOfAcks)
	if err != nil {
		return 0, err
	}

	endTimeMilli := time.Now().UnixMilli()

	// Step 5: If shouldVerifyTimeout is true : Assert that the WAIT command
	// returned after the timeout
	if waitTest.ShouldVerifyTimeout {
		threshold := 500 // ms
		elapsedTimeMilli := endTimeMilli - startTimeMilli
		client.Log(fmt.Sprintf("WAIT command returned after %v ms", elapsedTimeMilli))
		if math.Abs(float64(elapsedTimeMilli-int64(waitTest.WaitTimeoutMilli))) > float64(threshold) {
			return 0, fmt.Errorf("Expected WAIT to return exactly after %v ms timeout elapsed.", waitTest.WaitTimeoutMilli)
		}
	}
	return newReplicationOffset, nil
}
