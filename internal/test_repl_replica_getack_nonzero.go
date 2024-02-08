package internal

import (
	"fmt"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
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

	r := resp3.NewReader(conn)
	w := resp3.NewWriter(conn)

	master := FakeRedisMaster{
		Reader: r,
		Writer: w,
		Logger: logger,
	}

	err = master.Handshake()
	if err != nil {
		return err
	}

	// If I don't read ACK sent by redis replica, it's not buffered
	// I can easily read the next ACK in the next `stage`
	// So ideally, I can ignore the ACKs in the `SET`` stages, and then only check for explicit GETACKs

	err = master.GetAck(0) // 37
	if err != nil {
		return err
	}

	master.Send([]string{"PING"}) // 14
	// actualMessages, err := readRespMessages(r, logger)
	// fmt.Println(actualMessages)

	err = master.GetAck(51) // 37
	if err != nil {
		return err
	}

	master.Send([]string{"SET", "foo", "123"}) // 31
	// actualMessages, err := readRespMessages(r, logger)
	// fmt.Println(actualMessages)

	master.Send([]string{"SET", "bar", "456"}) // 31
	// actualMessages, err := readRespMessages(r, logger)
	// fmt.Println(actualMessages)

	err = master.GetAck(150)
	if err != nil {
		return err
	}

	listener.Close()
	return nil
}
