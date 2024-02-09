package internal

import (
	"fmt"
	"strings"

	testerutils "github.com/codecrafters-io/tester-utils"
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
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	conn2, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	conn3, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
	}

	client := NewRedisClient("localhost:6379")

	replica1 := NewFakeRedisReplica(conn1, logger)
	replica2 := NewFakeRedisReplica(conn2, logger)
	replica3 := NewFakeRedisReplica(conn3, logger)

	replicas := []FakeRedisReplica{*replica1, *replica2, *replica3}

	for i := 0; i < len(replicas); i++ {
		replica := replicas[i]
		err = replica.Handshake()
		if err != nil {
			return err
		}
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

	for j := 0; j < 3; j++ {
		replica := replicas[j]
		r := replica.Reader
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
				logger.Successf("Received %v on replica : %v.", strings.Join(cmd, " "), j+1)
			}
		}
	}

	conn1.Close()
	conn2.Close()
	conn3.Close()

	return nil
}
