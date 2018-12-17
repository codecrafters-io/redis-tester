package main

import "fmt"
import "github.com/go-redis/redis"
import "time"

import "math/rand"

// Tests 'GET, SET'
func testGetSet(logger *customLogger) error {
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

	randomKey := strings[rand.Intn(10)]
	randomValue := strings[rand.Intn(10)]

	logger.Debugf("Setting key %s to %s", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 0).Result()
	if err != nil {
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected 'OK', got %s", resp)
	}

	logger.Debugf("Getting key %s", randomKey)
	resp, err = client.Get(randomKey).Result()
	if err != nil {
		return err
	}

	if resp != randomValue {
		return fmt.Errorf("Expected %s, got %s", randomValue, resp)
	}

	client.Close()
	return nil
}
