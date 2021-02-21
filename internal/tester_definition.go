package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages: []testerutils.Stage{
		{
			Slug:     "anti-cheat-1",
			Title:    "Anti-cheat 1",
			TestFunc: antiCheatTest,
		},
	},
	ExecutableFileName: "spawn_redis_server.sh",
	Stages: []testerutils.Stage{
		{
			Slug:     "init",
			Title:    "Bind to a port",
			TestFunc: testBindToPort,
		},
		{
			Slug:     "ping-pong",
			Title:    "Respond to PING",
			TestFunc: testPingPongOnce,
		},
		{
			Slug:     "ping-pong-multiple",
			Title:    "Respond to multiple PINGs",
			TestFunc: testPingPongMultiple,
		},
		{
			Slug:     "concurrent-clients",
			Title:    "Handle concurrent clients",
			TestFunc: testPingPongConcurrent,
		},
		{
			Slug:     "echo",
			Title:    "Implement the ECHO command",
			TestFunc: testEcho,
		},
		{
			Slug:     "set_get",
			Title:    "SET & GET",
			TestFunc: testGetSet,
		},
		{
			Slug:     "expiry",
			Title:    "Expiry!",
			TestFunc: testExpiry,
		},
	},
}
