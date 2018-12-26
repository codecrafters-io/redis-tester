package main

func report(result StageRunnerResult, apiKey string) error {
	logger := getLogger(false, "[reporter] ")
	logger.Infoln("Submitting test results...")

	err := tellLeaderboard(apiKey, result.lastStageIndex)
	return err
}

func tellLeaderboard(apiKey string, stage int) error {
	return nil
}
