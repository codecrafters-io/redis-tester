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
		BeforeSendCommand: func(command string, args ...string) {
			if len(args) > 0 {
				stageHarness.Logger.Infof("%s$ redis-cli %s %s", logPrefix, command, strings.Join(args, " "))
			} else {
				stageHarness.Logger.Infof("%s$ redis-cli %s", logPrefix, command)
			}
		},
		BeforeSendValue: func(value resp_value.Value) {
			stageHarness.Logger.Infof("%s$ %s", logPrefix, value.FormattedString())
		},
		BeforeSendBytes: func(bytes []byte) {
			stageHarness.Logger.Debugf("%sSent bytes: %q", logPrefix, string(bytes))
		},
		AfterBytesReceived: func(bytes []byte) {
			stageHarness.Logger.Debugf("%sReceived bytes: %q", logPrefix, string(bytes))
		},
		AfterReadValue: func(value resp_value.Value) {
			stageHarness.Logger.Debugf("%sReceived RESP value: %s", logPrefix, value.FormattedString())
		},
	}
}

func NewFromAddr(stageHarness *test_case_harness.TestCaseHarness, addr string, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}
	return resp_connection.NewRespConnectionFromAddrWithCallbacks(
		addr, defaultCallbacks(stageHarness, logPrefix),
	)
}

func NewFromConn(stageHarness *test_case_harness.TestCaseHarness, conn net.Conn, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}

	return resp_connection.NewRespConnectionFromConnWithCallbacks(
		conn, defaultCallbacks(stageHarness, logPrefix),
	)
}
