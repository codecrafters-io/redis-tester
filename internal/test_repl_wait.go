package internal

import (
	"fmt"
	"strconv"
	"strings"
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

	client.SendAndAssert([]string{"SET", "foo", "123"}, []string{"OK"})

	err = client.Send([]string{"WAIT", "1", "500"})
	if err != nil {
		return err
	}

	logger.Infof("Start receiving on Replicas")
	ANSWER := 1
	// READ STREAM ON ALL REPLICAS
	for i := 0; i < replicaCount; i++ {
		replica := replicas[i]

		var skipFirstAssert bool
		skipFirstAssert = false
		actualMessages, err := readRespMessages(replica.Reader, logger)
		if strings.ToUpper(actualMessages[0]) != "SELECT" {
			skipFirstAssert = true
			expectedMessages := []string{"SET", "foo", "123"}
			offset += GetByteOffset(expectedMessages)
			err = assertMessages(actualMessages, expectedMessages, logger, true)
			if err != nil {
				return err
			}
		}

		if !skipFirstAssert {
			err, o := readAndAssertMessages(replica.Reader, []string{"SET", "foo", "123"}, logger, true)
			offset += o
			if err != nil {
				return err
			}
		}

		err, o := readAndAssertMessages(replica.Reader, []string{"REPLCONF", "GETACK", "*"}, logger, true)
		offset += o
		if err != nil {
			return err
		}
	}

	for i := 0; i < ANSWER; i++ {
		replica := replicas[i]

		msg := []string{"REPLCONF", "ACK", strconv.Itoa(offset)}
		err = replica.Writer.WriteCommand(msg...)
		if err != nil {
			return err
		}
		replica.Writer.Flush()
		replica.Logger.Infof("%s sent.", strings.ReplaceAll(strings.Join(msg, " "), "\r\n", ""))
	}

	readAndAssertIntMessage(client.Reader, ANSWER, logger)

	//////////////////////////////////////////////////////

	client.SendAndAssert([]string{"SET", "bar", "456"}, []string{"OK"})

	err = client.Send([]string{"WAIT", "3", "500"})
	if err != nil {
		return err
	}

	OLD_OFFSET := offset
	fmt.Println("Start receiving on Replicas")
	ANSWER = 3
	// READ STREAM ON ALL REPLICAS
	for i := 0; i < replicaCount; i++ {
		offset = OLD_OFFSET
		replica := replicas[i]

		// err, o := readAndAssertMessages(replica.Reader, []string{"SELECT", "0"}, logger)
		// offset += o

		err, o := readAndAssertMessages(replica.Reader, []string{"SET", "bar", "456"}, logger, true)
		offset += o

		err, o = readAndAssertMessages(replica.Reader, []string{"REPLCONF", "GETACK", "*"}, logger, false)
		offset += o
		if err != nil {
			return err
		}
	}

	for i := 0; i < ANSWER; i++ {
		replica := replicas[i]

		msg := []string{"REPLCONF", "ACK", strconv.Itoa(offset)}
		err = replica.Writer.WriteCommand(msg...)
		if err != nil {
			return err
		}
		replica.Writer.Flush()
		replica.Logger.Infof("%s sent.", strings.ReplaceAll(strings.Join(msg, " "), "\r\n", ""))
	}

	readAndAssertIntMessage(client.Reader, ANSWER, logger)

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

	OLD_OFFSET = offset
	logger.Infof("Start receiving on Replicas")
	// READ STREAM ON ALL REPLICAS
	for i := 0; i < replicaCount; i++ {
		offset = OLD_OFFSET
		replica := replicas[i]

		err, o := readAndAssertMessages(replica.Reader, []string{"SET", "baz", "789"}, logger, true)
		offset += o

		err, o = readAndAssertMessages(replica.Reader, []string{"REPLCONF", "GETACK", "*"}, logger, false)
		offset += o
		if err != nil {
			return err
		}
	}

	for i := 0; i < ANSWER; i++ {
		replica := replicas[i]

		msg := []string{"REPLCONF", "ACK", strconv.Itoa(offset)}
		err = replica.Writer.WriteCommand(msg...)
		if err != nil {
			return err
		}
		replica.Writer.Flush()
		replica.Logger.Infof("%s sent.", strings.ReplaceAll(strings.Join(msg, " "), "\r\n", ""))
	}

	readAndAssertIntMessage(client.Reader, ANSWER, logger)
	endTimeMilli := time.Now().UnixMilli()

	DELTA := 500 // ms
	timeElapsed := endTimeMilli - startTimeMilli
	logger.Infof("WAIT command returned after %v ms", timeElapsed)
	if timeElapsed > int64(TIMEOUT)+int64(DELTA) || timeElapsed < int64(TIMEOUT)-int64(DELTA) {
		return fmt.Errorf("Expected WAIT to return only after %v ms timeout elapsed.", TIMEOUT)
	}
	return nil
}
