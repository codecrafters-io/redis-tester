package internal

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

func testReplMasterCmdProp(stageHarness *testerutils.StageHarness) error {
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
	res, _, _ := r.ReadValue()
	message := res.SmartResult()
	respStr, _ := message.(string)
	respParts := strings.Split(respStr, " ")
	logger.Successf("PONG received.")

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")

	w.WriteCommand("REPLCONF", "listening-port", "6380")
	res, _, _ = r.ReadValue()
	message = res.SmartResult()
	respStr, _ = message.(string)
	respParts = strings.Split(respStr, " ")
	logger.Successf("OK received.")

	w.WriteCommand("PSYNC", "?", "-1")
	res, _, _ = r.ReadValue()
	message = res.SmartResult()
	respStr, _ = message.(string)
	respParts = strings.Split(respStr, " ")
	command := respParts[0]
	offset := respParts[2]

	if command != "FULLRESYNC" {
		return fmt.Errorf("Expected FULLRESYNC from Master, received %v", command)
	}
	logger.Successf("FULLRESYNC received.")
	if offset != "0" {
		return fmt.Errorf("Expected offset to be 0 from Master, received %v", offset)
	}
	logger.Successf("offset = 0 received.")

	reader := bufio.NewReader(conn)
	var data []byte
	timeout := 2 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeout))

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
		}
		data = append(data, b)
	}

	conn.SetReadDeadline(time.Time{})

	dataString := string(data)[6:]
	// First 6 chars are for RESP `$176\r\n` or similar.
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	err = decoder.Parse(processRedisObject)
	if err != nil {
		return fmt.Errorf("Error while parsing RDB file : %v", err)
	}
	logger.Successf("RDB file received from master.")

	key1, value1 := "foo", "123"
	key2, value2 := "bar", "456"
	key3, value3 := "baz", "789"

	logger.Debugf("Setting key %s to %s", key1, value1)
	client.Do("SET", key1, value1)
	logger.Debugf("Setting key %s to %s", key2, value2)
	client.Do("SET", key2, value2)
	logger.Debugf("Setting key %s to %s", key3, value3)
	client.Do("SET", key3, value3)

	var cmds [][]string
	conn.SetReadDeadline(time.Now().Add(timeout))

	for {
		req, err := Decode(reader)
		if err != nil {
			if err == io.EOF {
				continue
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			fmt.Println(err)
			break
		}

		if len(req.Array()) == 0 {
			continue
		}

		var cmd []string
		for _, v := range req.Array() {
			cmd = append(cmd, v.String())
		}
		cmds = append(cmds, cmd)
	}

	index := 0
	possibleSelect := cmds[index]
	if strings.ToUpper(possibleSelect[0]) == "SELECT" {
		index += 1
	}

	err = (compareStringSlices(cmds[index], []string{"SET", key1, value1}))
	if err != nil {
		return err
	}
	logger.Successf("Received %v", strings.Join(cmds[index], " "))
	index += 1

	err = (compareStringSlices(cmds[index], []string{"SET", key2, value2}))
	if err != nil {
		return err
	}
	logger.Successf("Received %v", strings.Join(cmds[index], " "))
	index += 1

	err = (compareStringSlices(cmds[index], []string{"SET", key3, value3}))
	if err != nil {
		return err
	}
	logger.Successf("Received %v", strings.Join(cmds[index], " "))

	conn.Close()
	return nil
}
