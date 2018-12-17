package main

import "fmt"
import "github.com/go-redis/redis"
import "time"

import "math/rand"

// Tests 'ECHO'
func testEcho(logger *customLogger) error {
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
