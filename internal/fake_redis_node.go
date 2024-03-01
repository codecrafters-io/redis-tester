package internal

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

type FakeRedisNode struct {
	Conn      net.Conn
	Reader    *resp3.Reader
	Writer    *resp3.Writer
	Logger    *logger.Logger
	LogPrefix string
}

func NewFakeRedisNode(conn net.Conn, logger *logger.Logger) *FakeRedisNode {
	return &FakeRedisNode{
		Conn:      conn,
		Reader:    resp3.NewReader(conn),
		Writer:    resp3.NewWriter(conn),
		Logger:    logger,
		LogPrefix: "",
	}
}

func (node FakeRedisNode) Log(message string) error {
	node.Logger.Infof(node.LogPrefix + message)
	return nil
}

func (node FakeRedisNode) Send(sendMessage []string) error {
	node.Logger.Infof(node.LogPrefix+"$ redis-cli %v", strings.Join(sendMessage, " "))
	err := node.Writer.WriteCommand(sendMessage...)
	if err != nil {
		return err
	}
	node.Writer.Flush()
	return nil
}

func (node FakeRedisNode) SendAndAssertStringArray(sendMessage []string, receiveMessage []string) error {
	err := node.Send(sendMessage)
	if err != nil {
		return err
	}
	_, err = node.readAndAssertMessages(receiveMessage, false)
	if err != nil {
		return err
	}
	return nil
}

func (node FakeRedisNode) SendAndAssertString(sendMessage []string, receiveMessage string, caseSensitiveMatch bool) error {
	err := node.Send(sendMessage)
	if err != nil {
		return err
	}
	err = node.readAndAssertMessage(receiveMessage, caseSensitiveMatch)
	if err != nil {
		return err
	}
	return nil
}

func (node FakeRedisNode) SendAndAssertInt(sendMessage []string, receiveMessage int) error {
	err := node.Send(sendMessage)
	if err != nil {
		return err
	}
	err = node.readAndAssertIntMessage(receiveMessage)
	if err != nil {
		return err
	}
	return nil
}

func (node FakeRedisNode) readRespMessages() ([]string, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return nil, e
	}
	message := resp.SmartResult()
	slice, ok := message.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected message received: %v", message)
	}
	return convertToStringArray(slice)
}

func (node FakeRedisNode) readRespString() (string, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return "", e
	}

	message := resp.SmartResult()
	slice, ok := message.(string)
	if !ok {
		return "", fmt.Errorf("Unexpected message received: %v", message)
	}

	if resp.Type != resp3.TypeSimpleString && resp.Type != resp3.TypeBlobString {
		return "", fmt.Errorf("Expected string, but received %q", message)
	}

	return slice, nil
}

func (node FakeRedisNode) readRespInt() (int, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return 0, e
	}
	message := resp.SmartResult()
	slice, ok := message.(int64)
	if !ok {
		return 0, fmt.Errorf("Unexpected message received: %v", message)
	}
	integer := int(slice)
	return integer, nil
}

func (node FakeRedisNode) readAndAssertMessagesWithSkip(messages []string, skipMessage string, caseSensitiveMatch bool) (int, error) {
	// Reads RESP message, skips assert if the first word matches with
	// skipMessage (case insensitive), reads next RESP runs match on it.
	actualMessages, err := node.readRespMessages()
	offset := 0
	if err != nil {
		return offset, err
	}
	if strings.EqualFold(actualMessages[0], skipMessage) {
		node.Logger.Successf(node.LogPrefix + strings.Join(actualMessages, " ") + " received.")
		offset += GetByteOffset(actualMessages)
		actualMessages, err = node.readRespMessages() // Read next message
		if err != nil {
			return offset, err
		}
	}

	offset += GetByteOffset(actualMessages)
	err = node.assertMessages(actualMessages, messages, caseSensitiveMatch)
	if err != nil {
		return offset, err
	}
	return offset, err
}

func (node FakeRedisNode) readAndAssertMessages(messages []string, caseSensitiveMatch bool) (int, error) {
	actualMessages, err := node.readRespMessages()
	offset := GetByteOffset(messages)
	if err != nil {
		return 0, err
	}
	// node.Logger.Errorf("ACTMSG : %v", actualMessages)
	err = node.assertMessages(actualMessages, messages, caseSensitiveMatch)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (node FakeRedisNode) readAndAssertMessagesWithOr(messages [][]string, caseSensitiveMatch bool) (int, error) {
	actualMessages, err := node.readRespMessages()
	offset := GetByteOffset(actualMessages)
	if err != nil {
		return 0, err
	}

	err = node.assertMessagesWithOr(actualMessages, messages, caseSensitiveMatch)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (node FakeRedisNode) assertMessages(actualMessages []string, expectedMessages []string, caseSensitiveMatch bool) error {
	err := compareStringSlices(actualMessages, expectedMessages, caseSensitiveMatch)
	if err != nil {
		return err
	}
	node.Logger.Successf(node.LogPrefix + strings.Join(actualMessages, " ") + " received.")
	return nil
}

func (node FakeRedisNode) assertMessagesWithOr(actualMessages []string, expectedMessages [][]string, caseSensitiveMatch bool) error {
	err := compareStringSlicesWithOr(actualMessages, expectedMessages, caseSensitiveMatch)
	if err != nil {
		return err
	}
	node.Logger.Successf(node.LogPrefix + strings.Join(actualMessages, " ") + " received.")
	return nil
}
func (node FakeRedisNode) readAndAssertMessage(expectedMessage string, caseSensitiveMatch bool) error {
	actualMessage, err := node.readRespString()
	if err != nil {
		return err
	}
	if strings.Contains(expectedMessage, " * ") {
		// Wildcard present, do a array comparison
		actualMessageParts := strings.Split(actualMessage, " ")
		expectedMessageParts := strings.Split(expectedMessage, " ")
		err = compareStringSlices(actualMessageParts, expectedMessageParts, caseSensitiveMatch)
	} else {
		var a, e string
		if caseSensitiveMatch {
			a, e = actualMessage, expectedMessage
		} else {
			a, e = strings.ToUpper(actualMessage), strings.ToUpper(expectedMessage)
		}
		if a != e {
			err = fmt.Errorf("Expected: '%v' and actual: '%v' messages don't match", expectedMessage, actualMessage)
		}
	}
	if err != nil {
		return err
	}
	node.Logger.Successf(node.LogPrefix + actualMessage + " received.")
	return nil
}

func (node FakeRedisNode) readAndAssertIntMessage(expectedMessage int) error {
	actualMessage, err := node.readRespInt()
	if err != nil {
		return err
	}
	if actualMessage != expectedMessage {
		err = fmt.Errorf("Expected: '%v' and actual: '%v' messages don't match", expectedMessage, actualMessage)
	}
	if err != nil {
		return err
	}
	node.Logger.Successf(node.LogPrefix + strconv.Itoa(actualMessage) + " received.")
	return nil
}

func GetByteOffset(args []string) int {
	offset := 0
	offset += 2 * (2*len(args) + 1)
	offset += (len(strconv.Itoa(len(args))) + 1)
	for _, arg := range args {
		msgLen := len(arg)
		offset += (len(strconv.Itoa(msgLen)) + 1)
		offset += (msgLen)
	}

	return offset
}

func readAndCheckRDBFile(reader *resp3.Reader) error {
	req, err := parseRESPCommandRDB(reader)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	if len(req.data) == 0 {
		return fmt.Errorf("Couldn't read data.")
	}
	dataString := string(req.data)
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	return decoder.Parse(processRedisObject)
}
