package instrumented_resp_connection

import (
	"net"

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

			commandWithArgs := append([]string{command}, args...)
			logger.Infof("%s %s", commandPrefix, quoteCLICommand(commandWithArgs))
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
			logger.Debugf("Received RESP %s: %s", value.Type, value.FormattedString())
		},
	}
}

func NewFromAddr(baseLogger *logger.Logger, addr string, connIdentifier string) (*InstrumentedRespConnection, error) {
	logger := baseLogger.Clone()
	logger.PushSecondaryPrefix(connIdentifier)
	c, err := resp_connection.NewRespConnectionFromAddr(addr, defaultCallbacks(logger))
	if err != nil {
		return nil, err
	}
	return &InstrumentedRespConnection{
		RespConnection: c,
		logger:         logger,
	}, nil
}

func NewFromConn(baseLogger *logger.Logger, conn net.Conn, connIdentifier string) (*InstrumentedRespConnection, error) {
	logger := baseLogger.Clone()
	logger.PushSecondaryPrefix(connIdentifier)
	c, err := resp_connection.NewRespConnectionFromConn(conn, defaultCallbacks(logger))
	if err != nil {
		return nil, err
	}
	return &InstrumentedRespConnection{
		RespConnection: c,
		logger:         logger,
	}, nil
}

func (c *InstrumentedRespConnection) GetIdentifier() string {
	return c.logger.GetLastSecondaryPrefix()
}

// GetLogger returns a new logger with added secondary prefix: connection's identifier
func (c *InstrumentedRespConnection) GetLogger() *logger.Logger {
	return c.logger
}

// UpdateBaseLogger updates the connection's logger
func (c *InstrumentedRespConnection) UpdateBaseLogger(l *logger.Logger) {
	newLogger := l.Clone()
	newLogger.PushSecondaryPrefix(c.GetIdentifier())
	c.logger = newLogger
	c.UpdateCallBacks(defaultCallbacks(c.logger))
}

func (c *InstrumentedRespConnection) SetReadValueInterceptor(transformer func(value resp_value.Value) resp_value.Value) {
	c.ReadValueInterceptor = transformer
}

func (c *InstrumentedRespConnection) UnsetReadValueInterceptor() {
	c.ReadValueInterceptor = nil
}
