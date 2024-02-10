package internal

import (
	"fmt"
	"strconv"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testWait(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	// replicas can be : [3, 5]
	replicaCount := testerutils_random.RandomInt(3, 5)
	offset := 0

	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	var replicas []*FakeRedisReplica
	for i := 0; i < replicaCount; i++ {
		logger.Debugf("Creating replica : %v.", i+1)
		conn, err := NewRedisConn("", "localhost:6379")
		defer conn.Close()
		if err != nil {
			fmt.Println("Error connecting to TCP server:", err)
		}

		replica := NewFakeRedisReplica(conn, logger)
		replicas = append(replicas, replica)
		replica.LogPrefix = fmt.Sprintf("[replica-%v] ", i+1)

		err = replica.Handshake()
		if err != nil {
			return err
		}
	}

	conn, err := NewRedisConn("", "localhost:6379")
	defer conn.Close()
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewFakeRedisMaster(conn, logger)
	client.LogPrefix = "[client] "

	client.SendAndAssert([]string{"SET", "foo", "123"}, []string{"OK"})

	err = client.Send([]string{"WAIT", "1", "500"})
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////

	ANSWER := 1
	// READ STREAM ON ALL REPLICAS
	for i := 0; i < replicaCount; i++ {
		replica := replicas[i]

		// Redis will send SELECT, but not expected from Users.
		err, _ = replica.readAndAssertMessagesWithSkip([]string{"SET", "foo", "123"}, "SELECT", true)
		if err != nil {
			return err
		}

		err, o := replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, true)
		offset += o
		if err != nil {
			return err
		}
	}

	for i := 0; i < ANSWER; i++ {
		replica := replicas[i]
		replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
	}

	client.readAndAssertIntMessage(ANSWER)

	//////////////////////////////////////////////////////

	client.SendAndAssert([]string{"SET", "baz", "789"}, []string{"OK"})

	ANSWER = min(replicaCount, testerutils_random.RandomInt(0, 6))
	TIMEOUT := 2000
	sendCount := strconv.Itoa(ANSWER + 1)
	startTimeMilli := time.Now().UnixMilli()
	err = client.Send([]string{"WAIT", sendCount, strconv.Itoa(TIMEOUT)})
	if err != nil {
		return err
	}

	OLD_OFFSET := offset
	// READ STREAM ON ALL REPLICAS
	for i := 0; i < replicaCount; i++ {
		offset = OLD_OFFSET
		replica := replicas[i]

		err, o := replica.readAndAssertMessages([]string{"SET", "baz", "789"}, true)
		offset += o

		err, o = replica.readAndAssertMessages([]string{"REPLCONF", "GETACK", "*"}, false)
		offset += o
		if err != nil {
			return err
		}
	}

	for i := 0; i < ANSWER; i++ {
		replica := replicas[i]
		replica.Send([]string{"REPLCONF", "ACK", strconv.Itoa(offset)})
	}

	client.readAndAssertIntMessage(ANSWER)

	//////////////////////////////////////////////////////

	endTimeMilli := time.Now().UnixMilli()

	DELTA := 500 // ms
	timeElapsed := endTimeMilli - startTimeMilli
	client.Log(fmt.Sprintf("WAIT command returned after %v ms", timeElapsed))
	if timeElapsed > int64(TIMEOUT)+int64(DELTA) || timeElapsed < int64(TIMEOUT)-int64(DELTA) {
		return fmt.Errorf("Expected WAIT to return only after %v ms timeout elapsed.", TIMEOUT)
	}
	return nil
}
