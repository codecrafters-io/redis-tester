package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/tester-utils/logger"
)

type FakeRedisReplica struct {
	FakeRedisNode
}

func NewFakeRedisReplica(conn net.Conn, logger *logger.Logger) *FakeRedisReplica {
	return &FakeRedisReplica{
		FakeRedisNode: *NewFakeRedisNode(conn, logger),
	}
}

func (replica FakeRedisReplica) SendAndAssertMessage(sendMessage []string, receiveMessage string, caseSensitiveMatch bool) error {
	err := replica.Send(sendMessage)
	if err != nil {
		return err
	}
	err = replica.readAndAssertMessage(receiveMessage, caseSensitiveMatch)
	if err != nil {
		return err
	}
	return nil
}

func (replica FakeRedisReplica) Ping() error {
	return replica.SendAndAssertMessage([]string{"PING"}, "PONG", false)
}
func (replica FakeRedisReplica) ReplConfPort() error {
	return replica.SendAndAssertMessage([]string{"REPLCONF", "listening-port", "6380"}, "OK", false)
}
func (replica FakeRedisReplica) Psync() error {
	return replica.SendAndAssertMessage([]string{"PSYNC", "?", "-1"}, "FULLRESYNC * 0", false)
}

func (replica FakeRedisReplica) ReceiveRDB() error {
	err := readAndCheckRDBFile(replica.Reader)
	if err != nil {
		return fmt.Errorf(replica.LogPrefix+"Error while parsing RDB file : %v", err)
	}
	replica.Logger.Successf(replica.LogPrefix + "Successfully received RDB file from master.")
	return nil
}

func (replica FakeRedisReplica) Handshake() error {
	err := replica.Ping()
	if err != nil {
		return err
	}

	err = replica.ReplConfPort()
	if err != nil {
		return err
	}

	err = replica.Psync()
	if err != nil {
		return err
	}

	err = replica.ReceiveRDB()
	if err != nil {
		return err
	}
	return nil
}
