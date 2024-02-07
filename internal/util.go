package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

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

func readAndCheckRDBFileUsingDecode(reader *resp3.Reader) error {
	req, err := parseRESPCommand(reader, true)
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

func readAndCheckRDBFileRESP3(reader *resp3.Reader) error {
	var rdb []byte
	for {
		byte, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
		}
		rdb = append(rdb, byte)
		if reader.Buffered() == 0 {
			break
		}
	}
	if len(rdb) == 0 {
		return fmt.Errorf("Couldn't read data.")
	}
	dataString := string(rdb)
	// n := len(rdb)
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	return decoder.Parse(processRedisObject)
}

func readAndCheckRDBFile(reader *bufio.Reader) error {
	var rdb []byte
	for {
		byte, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
		}
		rdb = append(rdb, byte)
		if reader.Buffered() == 0 {
			break
		}
	}
	if len(rdb) == 0 {
		return fmt.Errorf("Couldn't read data.")
	}
	dataString := string(rdb)
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	return decoder.Parse(processRedisObject)
}

func readAndCheckRDBFileNew(reader *bufio.Reader) error {
	req, err := reader.ReadString('\n') // Read marker
	marker := strings.Split(strings.TrimSpace(req), ":")[1]
	// markerLength := len(marker)
	var rdb []byte
	for {
		byte, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
		}
		rdb = append(rdb, byte)
		if strings.Contains(string(rdb), marker) {
			break
		}
	}
	rdbLenth := len(rdb)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	if len(rdb) == 0 {
		return fmt.Errorf("Couldn't read RDB file.")
	}
	// fmt.Println(string(rdb))
	dataString := string(rdb[:rdbLenth-40])
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	return decoder.Parse(processRedisObject)
}
