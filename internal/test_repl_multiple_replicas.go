package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/smallnest/resp3"
)

func testReplMultipleReplicas(stageHarness *testerutils.StageHarness) error {
	deleteRDBfile()
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	conn1, err := NewRedisConn("", "localhost:6379")
	conn2, err := NewRedisConn("", "localhost:6379")
	conn3, err := NewRedisConn("", "localhost:6379")

	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewRedisClient("localhost:6379")

	r1 := resp3.NewReader(conn1)
	w1 := resp3.NewWriter(conn1)
	r2 := resp3.NewReader(conn2)
	w2 := resp3.NewWriter(conn2)
	r3 := resp3.NewReader(conn3)
	w3 := resp3.NewWriter(conn3)

	logger.Infof("$ redis-cli PING")
	writers := []*resp3.Writer{w1, w2, w3}
	readers := []*resp3.Reader{r1, r2, r3}

	for i := 0; i < 3; i++ {
		r, w := readers[i], writers[i]
		w.WriteCommand("PING")
		actualMessage, err := readRespString(r, logger)
		if err != nil {
			return err
		}
		if actualMessage != "PONG" {
			return fmt.Errorf("Expected 'PONG', got %v", actualMessage)
		}
		logger.Successf("PONG received on replica : %v.", i+1)

		logger.Infof("$ redis-cli REPLCONF listening-port 6380")
		w.WriteCommand("REPLCONF", "listening-port", "6380")
		actualMessage, err = readRespString(r, logger)
		if err != nil {
			return err
		}
		if actualMessage != "OK" {
			return fmt.Errorf("Expected 'OK', got %v", actualMessage)
		}
		logger.Successf("OK received on replica : %v.", i+1)

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
		logger.Successf("FULLRESYNC received on replica : %v.", i+1)
		if offset != "0" {
			return fmt.Errorf("Expected Offset to be 0 from Master, got %v", offset)
		}
		logger.Successf("Offset = 0 received on replica : %v.", i+1)

		err = readAndCheckRDBFile(r)
		if err != nil {
			return fmt.Errorf("Error while parsing RDB file : %v", err)
		}
		logger.Successf("Successfully received RDB file from master on replica : %v.", i+1)
	}

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

	for j := 0; j < 3; j++ {
		r, _ := readers[j], writers[j]
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
				logger.Successf("Received %v on replica : %v.", strings.Join(cmd, " "), j+1)
			}
		}
	}

	conn1.Close()
	conn2.Close()
	conn3.Close()

	return nil
}
