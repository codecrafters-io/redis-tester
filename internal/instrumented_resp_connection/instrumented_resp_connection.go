package instrumented_resp_connection

import (
	"net"
	"strings"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/logger"
)

func defaultCallbacks(logger *logger.Logger, logPrefix string) resp_connection.RespConnectionCallbacks {
	return resp_connection.RespConnectionCallbacks{
		BeforeSendCommand: func(reusedConnection bool, command string, args ...string) {
			logger.PushSecondaryPrefix(logPrefix)
			defer logger.PopSecondaryPrefix()
			var commandPrefix string
			if reusedConnection {
				commandPrefix = ">"
			} else {
				commandPrefix = "$ redis-cli"
			}

			if len(args) > 0 {
				logger.Infof("%s %s %s", commandPrefix, command, strings.Join(args, " "))
			} else {
				logger.Infof("%s %s", commandPrefix, command)
			}
		},
		BeforeSendValue: func(value resp_value.Value) {
			logger.WithAdditionalSecondaryPrefix(logPrefix, func() {
				logger.Infof("Sent %s", value.FormattedString())
			})
		},
		BeforeSendBytes: func(bytes []byte) {
			logger.WithAdditionalSecondaryPrefix(logPrefix, func() {
				logger.Debugf("Sent bytes: %q", string(bytes))
			})
		},
		AfterBytesReceived: func(bytes []byte) {
			logger.WithAdditionalSecondaryPrefix(logPrefix, func() {
				logger.Debugf("Received bytes: %q", string(bytes))
			})
		},
		AfterReadValue: func(value resp_value.Value) {
			valueTypeLowerCase := strings.ReplaceAll(strings.ToLower(value.Type), "_", " ")
			if valueTypeLowerCase == "nil" {
				valueTypeLowerCase = "null bulk string"
			}
			logger.WithAdditionalSecondaryPrefix(logPrefix, func() {
				logger.Debugf("Received RESP %s: %s", valueTypeLowerCase, value.FormattedString())
			})

		},
	}
}

func NewFromAddr(logger *logger.Logger, addr string, connIdentifier string) (*resp_connection.RespConnection, error) {
	return resp_connection.NewRespConnectionFromAddr(
		addr, connIdentifier, defaultCallbacks(logger, connIdentifier),
	)
}

func NewFromConn(logger *logger.Logger, conn net.Conn, connIdentifier string) (*resp_connection.RespConnection, error) {
	return resp_connection.NewRespConnectionFromConn(
		conn, defaultCallbacks(logger, connIdentifier), connIdentifier,
	)
}
