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
			Number:                  1,
			Title:                   "Bind to a port",
			TestFunc:                testBindToPort,
			ShouldRunPreviousStages: true,
			Timeout:                 15 * time.Second,
		},
		{
			Slug:                    "ping-pong",
			Number:                  2,
			Title:                   "Respond to PING",
			TestFunc:                testPingPongOnce,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "ping-pong-multiple",
			Title:                   "Respond to multiple PINGs",
			Number:                  3,
			TestFunc:                testPingPongMultiple,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "concurrent-clients",
			Number:                  4,
			Title:                   "Handle concurrent clients",
			TestFunc:                testPingPongConcurrent,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "echo",
			Number:                  5,
			Title:                   "Implement the ECHO command",
			TestFunc:                testEcho,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "set_get",
			Number:                  6,
			Title:                   "Implement the SET & GET commands",
			TestFunc:                testGetSet,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "expiry",
			Number:                  7,
			Title:                   "Expiry",
			TestFunc:                testExpiry,
			ShouldRunPreviousStages: true,
		},
	},
}
