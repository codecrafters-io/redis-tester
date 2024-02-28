package instrumented_redis_client

import (
	"strings"

	redis_client "github.com/codecrafters-io/redis-tester/internal/redis_client"
	"github.com/codecrafters-io/redis-tester/internal/resp"
	testerutils "github.com/codecrafters-io/tester-utils"
)

func NewInstrumentedRedisClient(stageHarness *testerutils.StageHarness, addr string) (*redis_client.RedisClient, error) {
	return redis_client.NewRedisClientWithCallbacks(
		addr,
		redis_client.RedisClientCallbacks{
			OnSendCommand: func(command string, args ...string) {
				if len(args) > 0 {
					stageHarness.Logger.Infof("$ redis-cli %s %s", command, strings.Join(args, " "))
				} else {
					stageHarness.Logger.Infof("$ redis-cli %s", command)
				}
			},
			OnRawSend: func(bytes []byte) {
				stageHarness.Logger.Debugf("Sent bytes: %q", string(bytes))
			},
			OnRawRead: func(bytes []byte) {
				stageHarness.Logger.Debugf("Received bytes: %q", string(bytes))
			},
			OnValueRead: func(value resp.Value) {
				stageHarness.Logger.Debugf("Received RESP value: %s", value.FormattedString())
			},
		},
	)
}
