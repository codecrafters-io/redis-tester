package internal

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/tester-utils/logger"
)

type FakeRedisMaster struct {
	FakeRedisNode
}

func NewFakeRedisMaster(conn net.Conn, logger *logger.Logger) *FakeRedisMaster {
	return &FakeRedisMaster{
		FakeRedisNode: *NewFakeRedisNode(conn, logger),
	}
}

func (master FakeRedisMaster) Assert(receiveMessages []string, sendMessage string, caseSensitiveMatch bool) error {
	_, err := master.readAndAssertMessages(receiveMessages, caseSensitiveMatch)
	if err != nil {
		return err
	}
	_, err = master.Writer.WriteString(sendMessage)
	if err != nil {
		return err
	}
	master.Writer.Flush()
	master.Logger.Infof("%s sent.", strings.ReplaceAll(sendMessage, "\r\n", ""))
	return nil
}

func (master FakeRedisMaster) AssertWithOr(receiveMessages [][]string, sendMessage string, caseSensitiveMatch bool) error {
	_, err := master.readAndAssertMessagesWithOr(receiveMessages, caseSensitiveMatch)
	if err != nil {
		return err
	}
	_, err = master.Writer.WriteString(sendMessage)
	if err != nil {
		return err
	}
	master.Writer.Flush()
	master.Logger.Infof("%s sent.", strings.ReplaceAll(sendMessage, "\r\n", ""))
	return nil
}

func (master FakeRedisMaster) AssertPing() error {
	return master.Assert([]string{"PING"}, "+PONG\r\n", false)
}

func (master FakeRedisMaster) AssertReplConfPort() error {
	return master.Assert([]string{"REPLCONF", "listening-port", "6380"}, "+OK\r\n", false)
}

func (master FakeRedisMaster) AssertReplConfCapa() error {
	return master.AssertWithOr([][]string{{"REPLCONF", "capa", "*"}, {"REPLCONF", "capa", "*", "capa", "*"}}, "+OK\r\n", false)
}
func (master FakeRedisMaster) AssertPsync() error {
	id := RandomAlphanumericString(40)
	response := "+FULLRESYNC " + id + " 0\r\n"
	return master.Assert([]string{"PSYNC", "?", "-1"}, response, false)
}
func (master FakeRedisMaster) GetAck(offset int) error {
	return master.SendAndAssertStringArray([]string{"REPLCONF", "GETACK", "*"}, []string{"REPLCONF", "ACK", strconv.Itoa(offset)})
}

func (master FakeRedisMaster) Wait(needReplicas int, timeout time.Duration, maxReplicas int) error {
	if err := master.Send([]string{"WAIT", strconv.Itoa(needReplicas), strconv.Itoa(int(timeout.Milliseconds()))}); err != nil {
		return err
	}
	actualMessage, err := master.readRespInt()
	if err != nil {
		return err
	}

	if needReplicas > maxReplicas {
		if actualMessage != maxReplicas {
			return fmt.Errorf("Expected: '%v' and actual: '%v' messages don't match", maxReplicas, actualMessage)
		}
	} else {
		if actualMessage < needReplicas || actualMessage > maxReplicas {
			return fmt.Errorf("Expected: '[%v, %v]' and actual: '%v' messages don't match", needReplicas, maxReplicas, actualMessage)
		}
	}
	master.Logger.Successf(master.LogPrefix + strconv.Itoa(actualMessage) + " received.")
	return nil
}

func (master FakeRedisMaster) Handshake() error {
	err := master.AssertPing()
	if err != nil {
		return err
	}

	err = master.AssertReplConfPort()
	if err != nil {
		return err
	}

	err = master.AssertReplConfCapa()
	if err != nil {
		return err
	}

	err = master.AssertPsync()
	if err != nil {
		return err
	}

	response := SendRDBFile()
	master.Writer.Write(response)
	master.Logger.Infof("RDB file sent.")
	err = master.Writer.Flush()
	return err

}
