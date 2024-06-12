package instrumented_resp_connection

import (
	"net"
	"strings"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func defaultCallbacks(stageHarness *test_case_harness.TestCaseHarness, logPrefix string) resp_connection.RespConnectionCallbacks {
	return resp_connection.RespConnectionCallbacks{
		BeforeSendCommand: func(reusedConnection bool, command string, args ...string) {
			var commandPrefix string
			if reusedConnection {
				commandPrefix = ">"
			} else {
				commandPrefix = "$ redis-cli"
			}

			if len(args) > 0 {
				stageHarness.Logger.Infof("%s%s %s %s", logPrefix, commandPrefix, command, strings.Join(args, " "))
			} else {
				stageHarness.Logger.Infof("%s%s %s", logPrefix, commandPrefix, command)
			}
		},
		BeforeSendValue: func(value resp_value.Value) {
			stageHarness.Logger.Infof("%sSent %s", logPrefix, value.FormattedString())
		},
		BeforeSendBytes: func(bytes []byte) {
			stageHarness.Logger.Debugf("%sSent bytes: %q", logPrefix, string(bytes))
		},
		AfterBytesReceived: func(bytes []byte) {
			stageHarness.Logger.Debugf("%sReceived bytes: %q", logPrefix, string(bytes))
		},
		AfterReadValue: func(value resp_value.Value) {
			valueTypeLowerCase := strings.ReplaceAll(strings.ToLower(value.Type), "_", " ")
			if valueTypeLowerCase == "nil" {
				valueTypeLowerCase = "null bulk string"
			}
			stageHarness.Logger.Debugf("%sReceived RESP %s: %s", logPrefix, valueTypeLowerCase, value.FormattedString())

		},
	}
}

func NewFromAddr(stageHarness *test_case_harness.TestCaseHarness, addr string, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}
	return resp_connection.NewRespConnectionFromAddr(
		addr, defaultCallbacks(stageHarness, logPrefix),
	)
}

func NewFromConn(stageHarness *test_case_harness.TestCaseHarness, conn net.Conn, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}

	return resp_connection.NewRespConnectionFromConn(
		conn, defaultCallbacks(stageHarness, logPrefix),
	)
}
