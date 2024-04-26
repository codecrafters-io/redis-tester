package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

func IsSelectCommand(value resp_value.Value) bool {
	return value.Type == resp_value.ARRAY &&
		len(value.Array()) > 0 &&
		value.Array()[0].Type == resp_value.BULK_STRING &&
		strings.ToLower(value.Array()[0].String()) == "select"
}

func SpawnReplicas(replicaCount int, stageHarness *test_case_harness.TestCaseHarness, logger *logger.Logger, addr string) ([]*resp_connection.RespConnection, error) {
	var replicas []*resp_connection.RespConnection
	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}

	listeningPort := 6380
	for j := 0; j < replicaCount; j++ {
		logger.Debugf("Creating replica: %v", j+1)
		replica, err := instrumented_resp_connection.NewFromAddr(stageHarness, addr, fmt.Sprintf("replica-%v", j+1))
		if err != nil {
			logFriendlyError(logger, err)
			return nil, err
		}

		if err := sendHandshakeTestCase.RunAll(replica, logger, listeningPort); err != nil {
			return nil, err
		}
		listeningPort += 1
		// The bytes received and sent during the handshake don't count towards offset.
		// After finishing the handshake we reset the counters.
		replica.ResetByteCounters()

		replicas = append(replicas, replica)
	}
	return replicas, nil
}
