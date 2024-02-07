package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

type FakeRedisReplica struct {
	Reader *resp3.Reader
	Writer *resp3.Writer
	Logger *logger.Logger
}

func (replica FakeRedisReplica) Send(sendMessage []string, receiveMessage string) error {
	replica.Logger.Infof("$ redis-cli %v", strings.Join(sendMessage, " "))
	replica.Writer.WriteCommand(sendMessage...)
	err := readAndAssertMessage(replica.Reader, receiveMessage, replica.Logger)
	if err != nil {
		return err
	}
	return nil
}

func (replica FakeRedisReplica) Ping() error {
	return replica.Send([]string{"PING"}, "PONG")
}
func (replica FakeRedisReplica) ReplConfPort() error {
	return replica.Send([]string{"REPLCONF", "listening-port", "6380"}, "OK")
}
func (replica FakeRedisReplica) Psync() error {
	return replica.Send([]string{"PSYNC", "?", "-1"}, "FULLRESYNC * 0")
}

func (replica FakeRedisReplica) ReceiveRDB() error {
	err := readAndCheckRDBFile(replica.Reader)
	if err != nil {
		return fmt.Errorf("Error while parsing RDB file : %v", err)
	}
	replica.Logger.Successf("Successfully received RDB file from master.")
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

func readAndAssertMessages(reader *resp3.Reader, messages []string, logger *logger.Logger) error {
	actualMessages, err := readRespMessages(reader, logger)
	if err != nil {
		return err
	}
	expectedMessages := []string(messages)
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf(strings.Join(actualMessages, " ") + " received.")
	return nil
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

func sendAndLogMessage(writer *resp3.Writer, message string, logger *logger.Logger) error {
	if _, err := writer.WriteString(message); err != nil {
		return err
	}
	writer.Flush()
	logger.Infof("%s sent.", strings.ReplaceAll(message, "\r\n", ""))
	return nil
}
