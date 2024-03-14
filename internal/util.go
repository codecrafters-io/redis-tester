package internal

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

func SpawnReplicas(replicaCount int, stageHarness *test_case_harness.TestCaseHarness, logger *logger.Logger, addr string) ([]*resp_connection.RespConnection, error) {
	var replicas []*resp_connection.RespConnection
	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}

	for j := 0; j < replicaCount; j++ {
		logger.Debugf("Creating replica: %v", j+1)
		replica, err := instrumented_resp_connection.NewInstrumentedRespClient(stageHarness, addr, fmt.Sprintf("replica-%v", j+1))
		if err != nil {
			logFriendlyError(logger, err)
			return nil, err
		}

		if err := sendHandshakeTestCase.RunAll(replica, logger); err != nil {
			return nil, err
		}

		// Reset received and sent byte offset here.
		replica.InitializeOffset()
		replicas = append(replicas, replica)
	}
	return replicas, nil
}
