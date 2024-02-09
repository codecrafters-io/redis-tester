package internal

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

type FakeRedisMaster struct {
	Reader *resp3.Reader
	Writer *resp3.Writer
	Logger *logger.Logger
}

func NewFakeRedisMaster(conn net.Conn, logger *logger.Logger) *FakeRedisMaster {
	return &FakeRedisMaster{
		Reader: resp3.NewReader(conn),
		Writer: resp3.NewWriter(conn),
		Logger: logger,
	}
}

func (master FakeRedisMaster) Assert(receiveMessages []string, sendMessage string) error {
	err, _ := readAndAssertMessages(master.Reader, receiveMessages, master.Logger)
	_, err = master.Writer.WriteString(sendMessage)
	if err != nil {
		return err
	}
	master.Writer.Flush()
	master.Logger.Infof("%s sent.", strings.ReplaceAll(sendMessage, "\r\n", ""))
	return nil
}

func (master FakeRedisMaster) AssertPing() error {
	return master.Assert([]string{"PING"}, "+PONG\r\n")
}

func (master FakeRedisMaster) AssertReplConfPort() error {
	return master.Assert([]string{"REPLCONF", "listening-port", "6380"}, "+OK\r\n")
}

func (master FakeRedisMaster) AssertReplConfCapa() error {
	return master.Assert([]string{"REPLCONF", "*", "*", "*", "*"}, "+OK\r\n")
}
func (master FakeRedisMaster) AssertPsync() error {
	id, _ := RandomAlphanumericString(40)
	response := "+FULLRESYNC " + id + " 0\r\n"
	return master.Assert([]string{"PSYNC", "?", "-1"}, response)
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
	err, _ = readAndAssertMessages(master.Reader, receiveMessage, master.Logger)
	if err != nil {
		return err
	}
	return nil
}

func (master FakeRedisMaster) SendAndAssertString(sendMessage []string, receiveMessage string) error {
	err := master.Send(sendMessage)
	if err != nil {
		return err
	}
	err = readAndAssertMessage(master.Reader, receiveMessage, master.Logger)
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
	err = readAndAssertIntMessage(master.Reader, receiveMessage, master.Logger)
	if err != nil {
		return err
	}
	return nil
}

func (master FakeRedisMaster) Send(sendMessage []string) error {
	// Helps in logs.
	master.Logger.Infof("$ redis-cli %v", strings.Join(sendMessage, " "))
	err := master.Writer.WriteCommand(sendMessage...)
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

type FakeRedisReplica struct {
	Reader *resp3.Reader
	Writer *resp3.Writer
	Logger *logger.Logger
}

func NewFakeRedisReplica(conn net.Conn, logger *logger.Logger) *FakeRedisReplica {
	return &FakeRedisReplica{
		Reader: resp3.NewReader(conn),
		Writer: resp3.NewWriter(conn),
		Logger: logger,
	}
}

func (replica FakeRedisReplica) SendAndAssertMessage(sendMessage []string, receiveMessage string) error {
	replica.Logger.Infof("$ redis-cli %v", strings.Join(sendMessage, " "))
	replica.Writer.WriteCommand(sendMessage...)
	err := readAndAssertMessage(replica.Reader, receiveMessage, replica.Logger)
	if err != nil {
		return err
	}
	return nil
}

func (replica FakeRedisReplica) Ping() error {
	return replica.SendAndAssertMessage([]string{"PING"}, "PONG")
}
func (replica FakeRedisReplica) ReplConfPort() error {
	return replica.SendAndAssertMessage([]string{"REPLCONF", "listening-port", "6380"}, "OK")
}
func (replica FakeRedisReplica) Psync() error {
	return replica.SendAndAssertMessage([]string{"PSYNC", "?", "-1"}, "FULLRESYNC * 0")
}

func (replica FakeRedisReplica) ReceiveRDB() error {
	err := readAndCheckRDBFile(replica.Reader)
	if err != nil {
		return fmt.Errorf("Error while parsing RDB file : %v", err)
	}
	replica.Logger.Successf("Successfully received RDB file from master.")
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

func compareStringSlices(actual, expected []string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("Length mismatch between actual message and expected message.")
	}

	for i := range actual {
		// Wildcard for comparison.
		if expected[i] == "*" {
			continue
		}

		a, e := strings.ToUpper(actual[i]), strings.ToUpper(expected[i])
		if a != e {
			return fmt.Errorf("Expected : '%v' and actual : '%v' messages don't match", e, a)
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

func readRespMessages(reader *resp3.Reader, logger *logger.Logger) ([]string, error) {
	resp, b, e := reader.ReadValue()
	if e != nil {
		logger.Debugf(string(b))
		return nil, e
	}
	message := resp.SmartResult()
	slice, _ := message.([]interface{})
	return convertToStringArray(slice)
}

func readRespString(reader *resp3.Reader, logger *logger.Logger) (string, error) {
	resp, b, e := reader.ReadValue()
	if e != nil {
		logger.Debugf(string(b))
		return "", e
	}
	message := resp.SmartResult()
	slice, _ := message.(string)
	return slice, nil
}

func readRespInt(reader *resp3.Reader, logger *logger.Logger) (int, error) {
	resp, b, e := reader.ReadValue()
	if e != nil {
		logger.Debugf(string(b))
		return 0, e
	}
	message := resp.SmartResult()
	slice, err := message.(int64)
	if err != true {
		logger.Debugf("Unable to convert %v", message)
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

func readAndAssertMessages(reader *resp3.Reader, messages []string, logger *logger.Logger) (error, int) {
	actualMessages, err := readRespMessages(reader, logger)
	offset := GetByteOffset(messages)
	if err != nil {
		return err, 0
	}
	// fmt.Println("ACTMSG :", actualMessages)
	expectedMessages := []string(messages)
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err, 0
	}
	logger.Successf(strings.Join(actualMessages, " ") + " received.")
	return nil, offset
}

func readAndAssertMessage(reader *resp3.Reader, expectedMessage string, logger *logger.Logger) error {
	actualMessage, err := readRespString(reader, logger)
	if err != nil {
		return err
	}
	if strings.Contains(expectedMessage, " * ") {
		// Wildcard present, do a array comparison
		actualMessageParts := strings.Split(actualMessage, " ")
		expectedMessageParts := strings.Split(expectedMessage, " ")
		err = compareStringSlices(actualMessageParts, expectedMessageParts)
	} else {
		if actualMessage != expectedMessage {
			err = fmt.Errorf("Expected '%v', got '%v'", expectedMessage, actualMessage)
		}
	}
	if err != nil {
		return err
	}
	logger.Successf(actualMessage + " received.")
	return nil
}

func readAndAssertIntMessage(reader *resp3.Reader, expectedMessage int, logger *logger.Logger) error {
	actualMessage, err := readRespInt(reader, logger)
	if err != nil {
		return err
	}
	if actualMessage != expectedMessage {
		err = fmt.Errorf("Expected '%v', got '%v'", expectedMessage, actualMessage)
	}
	if err != nil {
		return err
	}
	logger.Successf(strconv.Itoa(actualMessage) + " received.")
	return nil
}

func sendAndLogMessage(writer *resp3.Writer, message string, logger *logger.Logger) error {
	if _, err := writer.WriteString(message); err != nil {
		return err
	}
	writer.Flush()
	logger.Infof("%s sent.", strings.ReplaceAll(message, "\r\n", ""))
	return nil
}

func RandomAlphanumericString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[charIndex.Int64()]
	}
	return string(result), nil
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
