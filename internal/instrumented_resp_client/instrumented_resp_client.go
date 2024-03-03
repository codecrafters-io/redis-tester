package instrumented_resp_client

import (
	"strings"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/client"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func NewInstrumentedRespClient(stageHarness *test_case_harness.TestCaseHarness, addr string, clientIdentifier string) (*resp_client.RespClient, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}

	return resp_client.NewRespClientWithCallbacks(
		addr,
		resp_client.RespClientCallbacks{
			BeforeSendCommand: func(command string, args ...string) {
				if len(args) > 0 {
					stageHarness.Logger.Infof("%s$ redis-cli %s %s", logPrefix, command, strings.Join(args, " "))
				} else {
					stageHarness.Logger.Infof("%s$ redis-cli %s", logPrefix, command)
				}
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
		},
	)
}
