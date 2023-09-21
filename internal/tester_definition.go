package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
	"time"
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
	},
}
