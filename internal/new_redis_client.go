package internal

import (
	"github.com/go-redis/redis"
	"time"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 10 * time.Second,
		MaxRetries:  20,
	})
}
