package internal

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

type FakeRedisNode struct {
	Reader    *resp3.Reader
	Writer    *resp3.Writer
	Logger    *logger.Logger
	LogPrefix string
}

func NewFakeRedisNode(conn net.Conn, logger *logger.Logger) *FakeRedisNode {
	return &FakeRedisNode{
		Reader:    resp3.NewReader(conn),
		Writer:    resp3.NewWriter(conn),
		Logger:    logger,
		LogPrefix: "",
	}
}

type FakeRedisMaster struct {
	FakeRedisNode
}

func NewFakeRedisMaster(conn net.Conn, logger *logger.Logger) *FakeRedisMaster {
	return &FakeRedisMaster{
		FakeRedisNode: *NewFakeRedisNode(conn, logger),
	}
}

type FakeRedisReplica struct {
	FakeRedisNode
}

func NewFakeRedisReplica(conn net.Conn, logger *logger.Logger) *FakeRedisReplica {
	return &FakeRedisReplica{
		FakeRedisNode: *NewFakeRedisNode(conn, logger),
	}
}

func (master FakeRedisMaster) Assert(receiveMessages []string, sendMessage string, caseSensitiveMatch bool) error {
	err, _ := master.readAndAssertMessages(receiveMessages, caseSensitiveMatch)
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
	return master.Assert([]string{"REPLCONF", "*", "*", "*", "*"}, "+OK\r\n", false)
}
func (master FakeRedisMaster) AssertPsync() error {
	id := RandomAlphanumericString(40)
	response := "+FULLRESYNC " + id + " 0\r\n"
	return master.Assert([]string{"PSYNC", "?", "-1"}, response, false)
}
func (master FakeRedisMaster) GetAck(offset int) error {
	return master.SendAndAssert([]string{"REPLCONF", "GETACK", "*"}, []string{"REPLCONF", "ACK", strconv.Itoa(offset)})
}

func (master FakeRedisMaster) Wait(replicas string, timeout string, expectedMessage int) error {
	return master.SendAndAssertInt([]string{"WAIT", replicas, timeout}, expectedMessage)
}

func (master FakeRedisMaster) SendAndAssert(sendMessage []string, receiveMessage []string) error {
	err := master.Send(sendMessage)
	if err != nil {
		return err
	}
	err, _ = master.readAndAssertMessages(receiveMessage, false)
	if err != nil {
		return err
	}
	return nil
}

func (master FakeRedisMaster) SendAndAssertString(sendMessage []string, receiveMessage string, caseSensitiveMatch bool) error {
	err := master.Send(sendMessage)
	if err != nil {
		return err
	}
	err = master.readAndAssertMessage(receiveMessage, caseSensitiveMatch)
	if err != nil {
		return err
	}
	return nil
}

func (master FakeRedisMaster) SendAndAssertInt(sendMessage []string, receiveMessage int) error {
	err := master.Send(sendMessage)
	if err != nil {
		return err
	}
	err = master.readAndAssertIntMessage(receiveMessage)
	if err != nil {
		return err
	}
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
func convertToStringArray(interfaceSlice []interface{}) ([]string, error) {
	stringSlice := make([]string, 0, len(interfaceSlice))

	for _, v := range interfaceSlice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element is not a string: %v", v)
		}
		stringSlice = append(stringSlice, str)
	}

	return stringSlice, nil
}

func compareStringSlices(actual, expected []string, caseSensitiveMatch bool) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("Length mismatch between actual message and expected message.")
	}

	for i := range actual {
		// Wildcard for comparison.
		if expected[i] == "*" {
			continue
		}
		var a, e string
		if caseSensitiveMatch {
			a, e = actual[i], expected[i]
		} else {
			// Case Insensitive matching
			a, e = strings.ToUpper(actual[i]), strings.ToUpper(expected[i])
		}
		if a != e {
			return fmt.Errorf("Expected: '%v' and actual: '%v' messages don't match", e, a)
		}
	}

	return nil
}

func parseInfoOutput(lines []string, seperator string) map[string]string {
	infoMap := make(map[string]string)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		parts := strings.Split(trimmedLine, seperator)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			infoMap[key] = value
		}
	}
	return infoMap
}

func (node FakeRedisNode) readRespMessages() ([]string, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return nil, e
	}
	message := resp.SmartResult()
	slice, _ := message.([]interface{})
	return convertToStringArray(slice)
}

func (node FakeRedisNode) readRespString() (string, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return "", e
	}
	message := resp.SmartResult()
	slice, _ := message.(string)
	return slice, nil
}

func (node FakeRedisNode) readRespInt() (int, error) {
	resp, b, e := node.Reader.ReadValue()
	if e != nil {
		node.Logger.Debugf(string(b))
		return 0, e
	}
	message := resp.SmartResult()
	slice, err := message.(int64)
	if err != true {
		node.Logger.Debugf("Unable to convert %v", message)
	}
	integer := int(slice)
	return integer, nil
}

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

// Used for parsing RDB file, to check validity.
func processRedisObject(o parser.RedisObject) bool {
	switch o.GetType() {
	case parser.StringType:
		str := o.(*parser.StringObject)
		println(str.Key, str.Value)
	case parser.ListType:
		list := o.(*parser.ListObject)
		println(list.Key, list.Values)
	case parser.HashType:
		hash := o.(*parser.HashObject)
		println(hash.Key, hash.Hash)
	case parser.ZSetType:
		zset := o.(*parser.ZSetObject)
		println(zset.Key, zset.Entries)
	}
	return true
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

func (node FakeRedisNode) readAndAssertMessagesWithSkip(messages []string, skipMessage string, caseSensitiveMatch bool) (error, int) {
	// Reads RESP message, skips assert if the first word matches with
	// skipMessage (case insensitive), reads next RESP runs match on it.
	actualMessages, err := node.readRespMessages()
	offset := 0
	if err != nil {
		return err, offset
	}
	if strings.ToUpper(actualMessages[0]) == strings.ToUpper(skipMessage) {
		node.Logger.Successf(node.LogPrefix + strings.Join(actualMessages, " ") + " received.")
		offset += GetByteOffset(actualMessages)
		actualMessages, err = node.readRespMessages() // Read next message
		if err != nil {
			return err, offset
		}
	}

	offset += GetByteOffset(actualMessages)
	err = node.assertMessages(actualMessages, messages, caseSensitiveMatch)
	if err != nil {
		return err, offset
	}
	return nil, offset
}

func (node FakeRedisNode) readAndAssertMessages(messages []string, caseSensitiveMatch bool) (error, int) {
	actualMessages, err := node.readRespMessages()
	offset := GetByteOffset(messages)
	if err != nil {
		return err, 0
	}
	// fmt.Println("ACTMSG :", actualMessages)
	err = node.assertMessages(actualMessages, messages, caseSensitiveMatch)
	if err != nil {
		return err, 0
	}
	return nil, offset
}

func (node FakeRedisNode) assertMessages(actualMessages []string, expectedMessages []string, caseSensitiveMatch bool) error {
	err := compareStringSlices(actualMessages, expectedMessages, caseSensitiveMatch)
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

func RandomAlphanumericString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		charIndex := testerutils_random.RandomInt(0, len(charset))
		result[i] = charset[charIndex]
	}
	return string(result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
