package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplMasterCmdProp(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	conn, err := NewRedisConn("", "localhost:6379")

	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewRedisClient("localhost:6379")

	r := resp3.NewReader(conn)
	w := resp3.NewWriter(conn)

	logger.Infof("$ redis-cli PING")

	w.WriteCommand("PING")
	actualMessage, err := readRespString(r, logger)
	if err != nil {
		return err
	}
	if actualMessage != "PONG" {
		return fmt.Errorf("Expected 'PONG', got %v", actualMessage)
	}
	logger.Successf("PONG received.")

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")
	w.WriteCommand("REPLCONF", "listening-port", "6380")
	actualMessage, err = readRespString(r, logger)
	if err != nil {
		return err
	}
	if actualMessage != "OK" {
		return fmt.Errorf("Expected 'OK', got %v", actualMessage)
	}
	logger.Successf("OK received.")

	logger.Infof("$ redis-cli PSYNC ? -1")
	w.WriteCommand("PSYNC", "?", "-1")
	actualMessage, err = readRespString(r, logger)
	if err != nil {
		return err
	}
	actualMessageParts := strings.Split(actualMessage, " ")
	command, offset := actualMessageParts[0], actualMessageParts[2]
	if command != "FULLRESYNC" {
		return fmt.Errorf("Expected 'FULLRESYNC' from Master, got %v", command)
	}
	logger.Successf("FULLRESYNC received.")
	if offset != "0" {
		return fmt.Errorf("Expected Offset to be 0 from Master, got %v", offset)
	}
	logger.Successf("Offset = 0 received.")

	err = readAndCheckRDBFile(r)
	if err != nil {
		return fmt.Errorf("Error while parsing RDB file : %v", err)
	}
	logger.Successf("Successfully received RDB file from master.")

	key1, value1 := "foo", "123"
	key2, value2 := "bar", "456"
	key3, value3 := "baz", "789"
	setMap := map[int][]string{
		1: {key1, value1},
		2: {key2, value2},
		3: {key3, value3},
	}
	for i := 1; i <= len(setMap); i++ { // We need order of commands preserved
		key, value := setMap[i][0], setMap[i][1]
		logger.Debugf("Setting key %s to %s", key, value)
		client.Do("SET", key, value)
	}

	i := 0
	for i < 3 {
		req, err := parseRESPCommand(r)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		var cmd []string
		for _, v := range req.Array() {
			cmd = append(cmd, v.String())
		}
		if len(cmd) > 0 && strings.ToUpper(cmd[0]) == "SET" {
			// User might not send SELECT, but Redis will send SELECT
			// Apart from SELECT we need 3 SETs
			i += 1
			key, value := setMap[i][0], setMap[i][1]
			err := compareStringSlices(cmd, []string{"SET", key, value})
			if err != nil {
				return err
			}
			logger.Successf("Received %v", strings.Join(cmd, " "))
		}
	}

	conn.Close()
	return nil
}
