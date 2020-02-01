package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
)

// Tests 'ECHO'
func testEcho(executable *Executable, logger *customLogger) error {
	b := NewRedisBinary(executable, logger)
	if err := b.Run(); err != nil {
		return err
	}
	defer b.Kill()

	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})

	strings := [10]string{
		"abcd",
		"defg",
		"heya",
		"heya",
		"heya",
		"heya",
		"heya",
		"heya",
		"heya",
		"heya",
	}

	randomString := strings[rand.Intn(10)]
	resp, err := client.Echo(randomString).Result()
	if err != nil {
		return err
	}

	if resp != randomString {
		return fmt.Errorf("Expected %s, got %s", randomString, resp)
	}

	client.Close()

	return nil
}
