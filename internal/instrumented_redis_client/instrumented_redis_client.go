package instrumented_redis_client

import (
	"strings"

	redis_client "github.com/codecrafters-io/redis-tester/internal/redis_client"
	"github.com/codecrafters-io/redis-tester/internal/resp"
	testerutils "github.com/codecrafters-io/tester-utils"
)

func NewInstrumentedRedisClient(stageHarness *testerutils.StageHarness, addr string, clientIdentifier string) (*redis_client.RedisClient, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}

	return redis_client.NewRedisClientWithCallbacks(
		addr,
		redis_client.RedisClientCallbacks{
			OnSendCommand: func(command string, args ...string) {
				if len(args) > 0 {
					stageHarness.Logger.Infof("%s$ redis-cli %s %s", logPrefix, command, strings.Join(args, " "))
				} else {
					stageHarness.Logger.Infof("%s$ redis-cli %s", logPrefix, command)
				}
			},
			OnRawSend: func(bytes []byte) {
				stageHarness.Logger.Debugf("%sSent bytes: %q", logPrefix, string(bytes))
			},
			OnRawRead: func(bytes []byte) {
				stageHarness.Logger.Debugf("%sReceived bytes: %q", logPrefix, string(bytes))
			},
			OnValueRead: func(value resp.Value) {
				stageHarness.Logger.Debugf("%sReceived RESP value: %s", logPrefix, value.FormattedString())
			},
		},
	)
}
