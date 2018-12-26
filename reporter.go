package main

func report(result StageRunnerResult) error {
	logger := getLogger(false, "[reporter] ")
	logger.Infoln("Submitting test results...")

	err := tellLeaderboard("abcd", result.lastStageIndex)
	return err
}

func tellLeaderboard(apiKey string, stage int) error {
	return nil
}
