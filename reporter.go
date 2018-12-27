package main

import "net/http"
import "bytes"
import "encoding/json"
import "fmt"
import "io/ioutil"

const apiURL = "https://redis-challenge-leaderboard.herokuapp.com"

func report(result StageRunnerResult, apiKey string) error {
	logger := getLogger(false, "[reporter] ")
	logger.Infoln("Submitting test results...")

	err := tellLeaderboard(apiKey, result.lastStageIndex, logger)
	return err
}

func tellLeaderboard(apiKey string, stage int, logger *customLogger) error {
	b, err := json.Marshal(map[string]interface{}{
		"api_key":                apiKey,
		"successful_stage_index": stage,
	})
	resp, err := http.Post(apiURL+"/report", "application/json", bytes.NewReader(b))
	if err != nil {
		logger.Errorf("Error when submitting test results: %s", err)
		return err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Error when reading response: %s", err)
		return err
	}

	if resp.StatusCode != 200 {
		logger.Errorf("Error when submitting tests results.")
		logger.Errorf("Response code: %d", resp.StatusCode)
		logger.Errorf("Body: %s", string(responseBody))
		return fmt.Errorf("error")
	}

	logger.Successln("Successfully reported tests results. Yay!")

	return nil
}
