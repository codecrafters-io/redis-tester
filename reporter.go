package main

func report(result StageRunnerResult) {
	logger := getLogger(false, "[reporter] ")
	logger.Infoln("Submitting test results...")
}
