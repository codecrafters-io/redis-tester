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

	// Logger is used to log a connection's network activity (sent/received)
	logger *logger.Logger
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

func NewFromAddr(logger *logger.Logger, addr string, connIdentifier string) (*InstrumentedRespConnection, error) {
	newLogger := logger.Clone()
	newLogger.PushSecondaryPrefix(connIdentifier)
	c, err := resp_connection.NewRespConnectionFromAddr(addr, defaultCallbacks(newLogger))
	if err != nil {
		return nil, err
	}
	return &InstrumentedRespConnection{
		RespConnection: c,
		logger:         newLogger,
	}, nil
}

func NewFromConn(logger *logger.Logger, conn net.Conn, connIdentifier string) (*InstrumentedRespConnection, error) {
	newLogger := logger.Clone()
	newLogger.PushSecondaryPrefix(connIdentifier)
	c, err := resp_connection.NewRespConnectionFromConn(conn, defaultCallbacks(newLogger))
	if err != nil {
		return nil, err
	}
	return &InstrumentedRespConnection{
		RespConnection: c,
		logger:         newLogger,
	}, nil
}

func (c *InstrumentedRespConnection) GetIdentifier() string {
	return c.logger.GetLastSecondaryPrefix()
}

// GetLogger returns a new logger with added secondary prefix: connection's identifier
func (c *InstrumentedRespConnection) GetLogger() *logger.Logger {
	return c.logger
}

// UpdateLogger updates the connection's logger
func (c *InstrumentedRespConnection) UpdateLogger(l *logger.Logger) {
	newLogger := l.Clone()
	newLogger.PushSecondaryPrefix(c.GetIdentifier())
	c.logger = newLogger
	c.UpdateCallBacks(defaultCallbacks(c.logger))
}
