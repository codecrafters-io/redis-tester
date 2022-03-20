package internal

import (
	"github.com/go-redis/redis"
	"net"
	"time"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 5 * time.Second,
		Dialer: func() (net.Conn, error) {
			attempts := 0

			for {
				var err error
				var conn net.Conn

				conn, err = net.Dial("tcp", "localhost:6379")

				if err == nil {
					return conn, nil
				}

				// Already a timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return nil, err
				}

				// 50 * 100ms = 5s
				if attempts > 50 {
					return nil, err
				}

				attempts += 1
				time.Sleep(100 * time.Millisecond)
			}
		},
	})
}
