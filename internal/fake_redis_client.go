package internal

import (
	"net"

	"github.com/codecrafters-io/tester-utils/logger"
)

type FakeRedisClient struct {
	FakeRedisNode
}

func NewFakeRedisClient(conn net.Conn, logger *logger.Logger) *FakeRedisClient {
	return &FakeRedisClient{
		FakeRedisNode: *NewFakeRedisNode(conn, logger),
	}
}
