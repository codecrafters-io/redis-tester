package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
	"time"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages: []testerutils.Stage{
		{
			Slug:                    "anti-cheat-1",
			Title:                   "Anti-cheat 1",
			TestFunc:                antiCheatTest,
			ShouldRunPreviousStages: true,
		},
	},
	ExecutableFileName: "spawn_redis_server.sh",
	Stages: []testerutils.Stage{
		{
			Slug:                    "init",
			Title:                   "Bind to a port",
			TestFunc:                testBindToPort,
			ShouldRunPreviousStages: true,
			Timeout:                 15 * time.Second,
		},
		{
			Slug:                    "ping-pong",
			Title:                   "Respond to PING",
			TestFunc:                testPingPongOnce,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "ping-pong-multiple",
			Title:                   "Respond to multiple PINGs",
			TestFunc:                testPingPongMultiple,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "concurrent-clients",
			Title:                   "Handle concurrent clients",
			TestFunc:                testPingPongConcurrent,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "echo",
			Title:                   "Implement the ECHO command",
			TestFunc:                testEcho,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "set_get",
			Title:                   "SET & GET",
			TestFunc:                testGetSet,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "expiry",
			Title:                   "Expiry!",
			TestFunc:                testExpiry,
			ShouldRunPreviousStages: true,
		},
	},
}
