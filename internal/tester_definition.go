package internal

import (
	"time"

	"github.com/codecrafters-io/tester-utils/tester_definition"
)

var testerDefinition = tester_definition.TesterDefinition{
	AntiCheatTestCases: []tester_definition.TestCase{
		{
			Slug:     "anti-cheat-1",
			TestFunc: antiCheatTest,
		},
	},
	ExecutableFileName:       "your_program.sh",
	LegacyExecutableFileName: "spawn_redis_server.sh",
	TestCases: []tester_definition.TestCase{
		// Base stages
		{
			Slug:     "jm1",
			TestFunc: testBindToPort,
			Timeout:  15 * time.Second,
		},
		{
			Slug:     "rg2",
			TestFunc: testPingPongOnce,
		},
		{
			Slug:     "wy1",
			TestFunc: testPingPongMultiple,
		},
		{
			Slug:     "zu2",
			TestFunc: testPingPongConcurrent,
		},
		{
			Slug:     "qq0",
			TestFunc: testEcho,
		},
		{
			Slug:     "la7",
			TestFunc: testGetSet,
		},
		{
			Slug:     "yz1",
			TestFunc: testExpiry,
		},
		// RDB Persistence
		{
			Slug:     "zg5",
			TestFunc: testRdbConfig,
		},
		{
			Slug:     "jz6",
			TestFunc: testRdbReadKey,
		},
		{
			Slug:     "gc6",
			TestFunc: testRdbReadStringValue,
		},
		{
			Slug:     "jw4",
			TestFunc: testRdbReadMultipleKeys,
		},
		{
			Slug:     "dq3",
			TestFunc: testRdbReadMultipleStringValues,
		},
		{
			Slug:     "sm4",
			TestFunc: testRdbReadValueWithExpiry,
		},
		{
			Slug:     "bw1",
			TestFunc: testReplBindToCustomPort,
		},
		{
			Slug:     "ye5",
			TestFunc: testReplInfo,
		},
		{
			Slug:     "hc6",
			TestFunc: testReplInfoReplica,
		},
		{
			Slug:     "xc1",
			TestFunc: testReplReplicationID,
		},
		{
			Slug:     "gl7",
			TestFunc: testReplReplicaSendsPing,
		},
		{
			Slug:     "eh4",
			TestFunc: testReplReplicaSendsReplconf,
		},
		{
			Slug:     "ju6",
			TestFunc: testReplReplicaSendsPsync,
		},
		{
			Slug:     "fj0",
			TestFunc: testReplMasterReplconf,
		},
		{
			Slug:     "vm3",
			TestFunc: testReplMasterPsync,
		},
		{
			Slug:     "cf8",
			TestFunc: testReplMasterPsyncRdb,
		},
		{
			Slug:     "zn8",
			TestFunc: testReplMasterCmdProp,
		},
		{
			Slug:     "hd5",
			TestFunc: testReplMultipleReplicas,
		},
		{
			Slug:     "yg4",
			TestFunc: testReplCmdProcessing,
		},
		{
			Slug:     "xv6",
			TestFunc: testReplGetaAckZero,
		},
		{
			Slug:     "yd3",
			TestFunc: testReplGetaAckNonZero,
		},
		{
			Slug:     "my8",
			TestFunc: testWaitZeroReplicas,
		},
		{
			Slug:     "tu8",
			TestFunc: testWaitZeroOffset,
		},
		{
			Slug:     "na2",
			TestFunc: testWait,
		},
		// Streams
		{
			Slug:     "cc3",
			TestFunc: testStreamsType,
		},
		{
			Slug:     "cf6",
			TestFunc: testStreamsXadd,
		},
		{
			Slug:     "hq8",
			TestFunc: testStreamsXaddValidateID,
		},
		{
			Slug:     "yh3",
			TestFunc: testStreamsXaddPartialAutoid,
		},
		{
			Slug:     "xu6",
			TestFunc: testStreamsXaddFullAutoid,
		},
		{
			Slug:     "zx1",
			TestFunc: testStreamsXrange,
		},
		{
			Slug:     "yp1",
			TestFunc: testStreamsXrangeMinID,
		},
		{
			Slug:     "fs1",
			TestFunc: testStreamsXrangeMaxID,
		},
		{
			Slug:     "um0",
			TestFunc: testStreamsXread,
		},
		{
			Slug:     "ru9",
			TestFunc: testStreamsXreadMultiple,
		},
		{
			Slug:     "bs1",
			TestFunc: testStreamsXreadBlock,
		},
		{
			Slug:     "hw1",
			TestFunc: testStreamsXreadBlockNoTimeout,
		},
		{
			Slug:     "xu1",
			TestFunc: testStreamsXreadBlockMaxID,
		},
		// Transactions
		{
			Slug:     "si4",
			TestFunc: testTxIncr1,
		},
		{
			Slug:     "lz8",
			TestFunc: testTxIncr2,
		},
		{
			Slug:     "mk1",
			TestFunc: testTxIncr3,
		},
		{
			Slug:     "pn0",
			TestFunc: testTxMulti,
		},
		{
			Slug:     "lo4",
			TestFunc: testTxExec,
		},
		{
			Slug:     "we1",
			TestFunc: testTxEmpty,
		},
		{
			Slug:     "rs9",
			TestFunc: testTxQueue,
		},
		{
			Slug:     "fy6",
			TestFunc: testTxSuccess,
		},
		{
			Slug:     "rl9",
			TestFunc: testTxDiscard,
		},
		{
			Slug:     "sg9",
			TestFunc: testTxErr,
		},
		{
			Slug:     "jf8",
			TestFunc: testTxMultiTx,
		},
		// Lists
		{
			Slug:     "mh6",
			TestFunc: testListRpush1,
		},
		{
			Slug:     "tn7",
			TestFunc: testListRpush2,
		},
		{
			Slug:     "lx4",
			TestFunc: testListRpush3,
		},
		{
			Slug:     "sf6",
			TestFunc: testListLrangePosIdx,
		},
		{
			Slug:     "ri1",
			TestFunc: testListLrangeNegIndex,
		},
		{
			Slug:     "gu5",
			TestFunc: testListLpush,
		},
		{
			Slug:     "fv6",
			TestFunc: testListLlen,
		},
		{
			Slug:     "ef1",
			TestFunc: testListLpop1,
		},
		{
			Slug:     "jp1",
			TestFunc: testListLpop2,
		},
		{
			Slug:     "ec3",
			TestFunc: testListBlpopNoTimeout,
		},
		{
			Slug:     "xj7",
			TestFunc: testListBlpopWithTimeout,
		},
		// Pub-Sub
		{
			Slug:     "mx3",
			TestFunc: testPubSubSubscribe1,
		},
		{
			Slug:     "zc8",
			TestFunc: testPubSubSubscribe2,
		},
		{
			Slug:     "aw8",
			TestFunc: testPubSubSubscribe3,
		},
		{
			Slug:     "lf1",
			TestFunc: testPubSubSubscribe4,
		},
		{
			Slug:     "hf2",
			TestFunc: testPubSubPublish1,
		},
		{
			Slug:     "dn4",
			TestFunc: testPubSubPublish2,
		},
		{
			Slug:     "ze9",
			TestFunc: testPubSubUnsubscribe,
		},
	},
}
