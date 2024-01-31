package internal

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/hdt3213/rdb/parser"
	"github.com/smallnest/resp3"
)

func testReplMasterPsyncRdb(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient()

	logger.Infof("$ redis-cli PING")
	resp, err := client.Do("PING").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "PONG" {
		return fmt.Errorf("Expected OK from Master, received %v", resp)
	}
	logger.Successf("PONG received.")

	logger.Infof("$ redis-cli REPLCONF listening-port 6380")
	resp, err = client.Do("REPLCONF", "listening-port", "6380").Result()

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected OK from Master, received %v", resp)
	}
	logger.Successf("OK received.")

	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}
	defer conn.Close()

	r := resp3.NewReader(conn)
	w := resp3.NewWriter(conn)

	w.WriteCommand("PSYNC", "?", "-1")
	response, _, _ := r.ReadValue()
	message := response.SmartResult()
	respStr, _ := message.(string)
	respParts := strings.Split(respStr, " ")
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
	timeout := 5 * time.Second

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))

		b, err := reader.ReadByte()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
		}
		data = append(data, b)
	}

	dataString := string(data)[6:]
	// First 6 chars are for RESP `$176\r\n` or similar.
	stringIOReader := strings.NewReader(dataString)
	decoder := parser.NewDecoder(stringIOReader)
	err = decoder.Parse(processRedisObject)
	if err != nil {
		return fmt.Errorf("Error while parsing RDB file : %v", err)
	}
	logger.Successf("RDB file received from master.")
	client.Close()
	return nil
}
