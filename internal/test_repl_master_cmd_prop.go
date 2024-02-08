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

	replica := FakeRedisReplica{
		Reader: r,
		Writer: w,
		Logger: logger,
	}

	err = replica.Handshake()
	if err != nil {
		return err
	}

	kvMap := map[int][]string{
		1: {"foo", "123"},
		2: {"bar", "456"},
		3: {"baz", "789"},
	}
	for i := 1; i <= len(kvMap); i++ { // We need order of commands preserved
		key, value := kvMap[i][0], kvMap[i][1]
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
			key, value := kvMap[i][0], kvMap[i][1]
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
