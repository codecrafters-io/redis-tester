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
			var commandPrefix string
			if reusedConnection {
				commandPrefix = ">"
			} else {
				commandPrefix = "$ redis-cli"
			}

			if len(args) > 0 {
				logger.Infof("%s%s %s %s", logPrefix, commandPrefix, command, strings.Join(args, " "))
			} else {
				logger.Infof("%s%s %s", logPrefix, commandPrefix, command)
			}
		},
		BeforeSendValue: func(value resp_value.Value) {
			logger.Infof("%sSent %s", logPrefix, value.FormattedString())
		},
		BeforeSendBytes: func(bytes []byte) {
			logger.Debugf("%sSent bytes: %q", logPrefix, string(bytes))
		},
		AfterBytesReceived: func(bytes []byte) {
			logger.Debugf("%sReceived bytes: %q", logPrefix, string(bytes))
		},
		AfterReadValue: func(value resp_value.Value) {
			valueTypeLowerCase := strings.ReplaceAll(strings.ToLower(value.Type), "_", " ")
			if valueTypeLowerCase == "nil" {
				valueTypeLowerCase = "null bulk string"
			}
			logger.Debugf("%sReceived RESP %s: %s", logPrefix, valueTypeLowerCase, value.FormattedString())

		},
	}
}

func NewFromAddr(logger *logger.Logger, addr string, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}
	return resp_connection.NewRespConnectionFromAddr(
		addr, defaultCallbacks(logger, logPrefix),
	)
}

func NewFromConn(logger *logger.Logger, conn net.Conn, clientIdentifier string) (*resp_connection.RespConnection, error) {
	logPrefix := ""
	if clientIdentifier != "" {
		logPrefix = clientIdentifier + ": "
	}

	return resp_connection.NewRespConnectionFromConn(
		conn, defaultCallbacks(logger, logPrefix),
	)
}
