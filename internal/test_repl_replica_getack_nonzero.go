package internal

import (
	"fmt"
	"math/rand"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplGetaAckNonZero(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
	}

	logger := stageHarness.Logger

	logger.Infof("Server is running on port 6379")

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}

	if err := replica.Run(); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}

	master := NewFakeRedisMaster(conn, logger)

	err = master.Handshake()
	if err != nil {
		return err
	}

	// If I don't read ACK sent by redis replica, it's not buffered
	// I can easily read the next ACK in the next `stage`
	// So ideally, I can ignore the ACKs in the `SET`` stages,
	// and then only check for explicit GETACKs

	offset := 0
	err = master.GetAck(offset) // 37
	if err != nil {
		return err
	}
	offset += GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	cmd := []string{"PING"}
	master.Send(cmd) // 14
	// actualMessages, err := readRespMessages(r, logger)
	// fmt.Println(actualMessages)
	offset += GetByteOffset(cmd)

	err = master.GetAck(offset) // 37
	if err != nil {
		return err
	}
	offset += GetByteOffset([]string{"REPLCONF", "GETACK", "*"})

	key, _ := RandomAlphanumericString(3 + rand.Intn(20))
	value, _ := RandomAlphanumericString(3 + rand.Intn(20))
	cmd = []string{"SET", key, value}
	master.Send(cmd) // 31
	// actualMessages, err = readRespMessages(r, logger)
	// fmt.Println(actualMessages)
	offset += GetByteOffset(cmd)

	key, _ = RandomAlphanumericString(3 + rand.Intn(20))
	value, _ = RandomAlphanumericString(3 + rand.Intn(20))
	cmd = []string{"SET", key, value}
	master.Send(cmd) // 31
	// actualMessages, err = readRespMessages(r, logger)
	// fmt.Println(actualMessages)
	offset += GetByteOffset(cmd)

	err = master.GetAck(offset)
	if err != nil {
		return err
	}

	listener.Close()
	return nil
}
