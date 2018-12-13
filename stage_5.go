package main

import "fmt"
import "github.com/go-redis/redis"
import "time"

import "math/rand"

// Tests Expiry
func runStage5(logger *customLogger) error {
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
	resp, err := client.Set(randomKey, randomValue, 100*time.Millisecond).Result()
	if err != nil {
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("Expected 'OK', got %s", resp)
	}

	resp, err = client.Get(randomKey).Result()
	if err != nil {
		return err
	}
	if resp != randomValue {
		return fmt.Errorf("Expected %s, got %s", randomValue, resp)
	}

	time.Sleep(101 * time.Millisecond)

	resp, err = client.Get(randomKey).Result()
	if err != redis.Nil {
		if err == nil {
			return fmt.Errorf("Expected nil, got %s", resp)
		}

		return err
	}

	client.Close()
	return nil
}
