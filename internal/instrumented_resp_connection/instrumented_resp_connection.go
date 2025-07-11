package instrumented_resp_connection

import (
	"net"
	"strings"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/logger"
)

type InstrumentedRespConnection struct {
	*resp_connection.RespConnection

	Logger *logger.Logger
}

func defaultCallbacks(logger *logger.Logger) resp_connection.RespConnectionCallbacks {
	return resp_connection.RespConnectionCallbacks{
		BeforeSendCommand: func(reusedConnection bool, command string, args ...string) {
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
			logger.Infof("Sent %s", value.FormattedString())
		},
		BeforeSendBytes: func(bytes []byte) {
			logger.Debugf("Sent bytes: %q", string(bytes))
		},
		AfterBytesReceived: func(bytes []byte) {
			logger.Debugf("Received bytes: %q", string(bytes))
		},
		AfterReadValue: func(value resp_value.Value) {
			valueTypeLowerCase := strings.ReplaceAll(strings.ToLower(value.Type), "_", " ")
			if valueTypeLowerCase == "nil" {
				valueTypeLowerCase = "null bulk string"
			}
			logger.Debugf("Received RESP %s: %s", valueTypeLowerCase, value.FormattedString())
		},
	}
}

func NewFromAddr(baseLogger *logger.Logger, addr string, connIdentifier string) (*resp_connection.RespConnection, error) {
	logger := baseLogger.Clone()
	logger.PushSecondaryPrefix(connIdentifier)

	return resp_connection.NewRespConnectionFromAddr(
		addr, connIdentifier, defaultCallbacks(logger),
	)
}

func NewFromConn(baseLogger *logger.Logger, conn net.Conn, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logger := baseLogger.Clone()
	logger.PushSecondaryPrefix(clientIdentifier)

	return resp_connection.NewRespConnectionFromConn(
		conn, defaultCallbacks(logger),
	)
}
