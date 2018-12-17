package main

import "fmt"
import "github.com/go-redis/redis"
import "time"

import "math/rand"

// Tests Expiry
func testExpiry(logger *customLogger) error {
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

	logger.Debugf("Setting key %s to %s, with expiry of 100ms", randomKey, randomValue)
	resp, err := client.Set(randomKey, randomValue, 100*time.Millisecond).Result()
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

	logger.Debugf("Sleeping for 101ms")
	time.Sleep(101 * time.Millisecond)

	logger.Debugf("Fetching value for key %s", randomKey)
	resp, err = client.Get(randomKey).Result()
	if err != redis.Nil {
		if err == nil {
			return fmt.Errorf("Expected nil, got %v", resp)
		}

		return err
	}

	client.Close()
	return nil
}
