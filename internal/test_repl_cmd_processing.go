package internal

import (
	"fmt"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplCmdProcessing(stageHarness *testerutils.StageHarness) error {
	master := NewRedisBinary(stageHarness)
	master.args = []string{
		"--port", "6379",
	}

	if err := master.Run(); err != nil {
		return err
	}

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}

	if err := replica.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	masterAddr, replicaAddr := "localhost:6379", "localhost:6380"
	masterClient := NewRedisClient(masterAddr)
	replicaClient := NewRedisClient(replicaAddr)

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
		masterClient.Do("SET", key, value)
	}

	for i := 1; i <= len(setMap); i++ {
		key, value := setMap[i][0], setMap[i][1]
		logger.Debugf("Getting key %s", key)
		resp, err := replicaClient.Get(key).Result()
		if err != nil {
			return err
		}
		if resp != value {
			return fmt.Errorf("Expected %#v, got %#v", value, resp)
		}
		logger.Successf("Received %v", resp)
	}

	masterClient.Close()
	replicaClient.Close()

	return nil
}
