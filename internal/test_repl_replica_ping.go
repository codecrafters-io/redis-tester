package internal

import (
	"fmt"
	"net"

	testerutils "github.com/codecrafters-io/tester-utils"
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
		if actual[i] != expected[i] {
			return fmt.Errorf("Expected : '%v' and actual : '%v' messages don't match", expected[i], actual[i])

		}
	}

	return nil
}

func testReplReplicaSendsPing(stageHarness *testerutils.StageHarness) error {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
	}
	defer listener.Close()
	logger := stageHarness.Logger

	logger.Infof("Server is running on port 6379.")

	replica := NewRedisBinary(stageHarness)
	replica.args = []string{
		"--port", "6380",
		"--replicaof", "localhost", "6379",
	}

	if err := replica.Run(); err != nil {
		return err
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		return err
	}

	r := resp3.NewReader(conn)

	resp, _, _ := r.ReadValue()
	message := resp.SmartResult()
	slice, _ := message.([]interface{})
	actualMessages, _ := convertToStringArray(slice)
	expectedMessages := []string{"PING"}
	err = compareStringSlices(actualMessages, expectedMessages)
	if err != nil {
		return err
	}
	logger.Successf("PING received.")
	return nil
}
