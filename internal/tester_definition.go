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
	},
}
