package internal

import (
	"net"
	"time"

	"github.com/go-redis/redis"
)

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        addr,
		DialTimeout: 2 * time.Second,
		Dialer: func() (net.Conn, error) {
			attempts := 0

			for {
				var err error
				var conn net.Conn

				conn, err = net.Dial("tcp", addr)

				if err == nil {
					return conn, nil
				}

				// Already a timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return nil, err
				}

				// 20 * 100ms = 2s
				if attempts > 20 {
					return nil, err
				}

				attempts += 1
				time.Sleep(100 * time.Millisecond)
			}
		},
	})
}
