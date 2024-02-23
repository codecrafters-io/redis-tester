package internal

import (
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatTestCases: []testerutils.TestCase{
		{
			Slug:     "anti-cheat-1",
			TestFunc: antiCheatTest,
		},
	},
	ExecutableFileName: "spawn_redis_server.sh",
	TestCases: []testerutils.TestCase{
		{
			Slug:     "init",
			TestFunc: testBindToPort,
			Timeout:  15 * time.Second,
		},
		{
			Slug:     "ping-pong",
			TestFunc: testPingPongOnce,
		},
		{
			Slug:     "ping-pong-multiple",
			TestFunc: testPingPongMultiple,
		},
		{
			Slug:     "concurrent-clients",
			TestFunc: testPingPongConcurrent,
		},
		{
			Slug:     "echo",
			TestFunc: testEcho,
		},
		{
			Slug:     "set_get",
			TestFunc: testGetSet,
		},
		{
			Slug:     "expiry",
			TestFunc: testExpiry,
		},
		{
			Slug:     "rdb-config",
			TestFunc: testRdbConfig,
		},
		{
			Slug:     "rdb-read-key",
			TestFunc: testRdbReadKey,
		},
		{
			Slug:     "rdb-read-string-value",
			TestFunc: testRdbReadStringValue,
		},
		{
			Slug:     "rdb-read-multiple-keys",
			TestFunc: testRdbReadMultipleKeys,
		},
		{
			Slug:     "rdb-read-multiple-string-values",
			TestFunc: testRdbReadMultipleStringValues,
		},
		{
			Slug:     "rdb-read-value-with-expiry",
			TestFunc: testRdbReadValueWithExpiry,
		},
		{
			Slug:     "repl-custom-port",
			TestFunc: testReplBindToCustomPort,
		},
		{
			Slug:     "repl-info",
			TestFunc: testReplInfo,
		},
		{
			Slug:     "repl-info-replica",
			TestFunc: testReplInfoReplica,
		},
		{
			Slug:     "repl-id",
			TestFunc: testReplReplicationID,
		},
		{
			Slug:     "repl-replica-ping",
			TestFunc: testReplReplicaSendsPing,
		},
		{
			Slug:     "repl-replica-replconf",
			TestFunc: testReplReplicaSendsReplconf,
		},
		{
			Slug:     "repl-replica-psync",
			TestFunc: testReplReplicaSendsPsync,
		},
		{
			Slug:     "repl-master-replconf",
			TestFunc: testReplMasterReplconf,
		},
		{
			Slug:     "repl-master-psync",
			TestFunc: testReplMasterPsync,
		},
		{
			Slug:     "repl-master-psync-rdb",
			TestFunc: testReplMasterPsyncRdb,
		},
		{
			Slug:     "repl-master-cmd-prop",
			TestFunc: testReplMasterCmdProp,
		},
		{
			Slug:     "repl-multiple-replicas",
			TestFunc: testReplMultipleReplicas,
		},
		{
			Slug:     "repl-cmd-processing",
			TestFunc: testReplCmdProcessing,
		},
		{
			Slug:     "repl-replica-getack",
			TestFunc: testReplGetaAckZero,
		},
		{
			Slug:     "repl-replica-getack-nonzero",
			TestFunc: testReplGetaAckNonZero,
		},
		{
			Slug:     "repl-wait-zero-replicas",
			TestFunc: testWaitZeroReplicas,
		},
		{
			Slug:     "repl-wait-zero-offset",
			TestFunc: testWaitZeroOffset,
		},
		{
			Slug:     "repl-wait",
			TestFunc: testWait,
		},
		{
			Slug:     "streams-type",
			TestFunc: testStreamsType,
		},
		{
			Slug:     "streams-xadd",
			TestFunc: testStreamsXadd,
		},
		{
			Slug:     "streams-xrange-max-id",
			TestFunc: testStreamsXrangeMaxId,
		},
	},
}
