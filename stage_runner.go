package main

import "fmt"
import "log"
import "os"

// StageRunnerResult is returned from StageRunner.Run()
type StageRunnerResult struct {
	failedAtStage Stage
	error         error
}

// IsSuccess says whether a StageRunnerResult was successful
// or not
func (res StageRunnerResult) IsSuccess() bool {
	return res.error == nil
}

// StageRunner is used to run multiple stages
type StageRunner struct {
	stages map[string]Stage
}

func newStageRunner() StageRunner {
	return StageRunner{
		stages: map[string]Stage{
			"stage-1": getStageOne(),
		},
	}
}

// Run tests in a specific StageRunner
func (r StageRunner) Run() StageRunnerResult {
	for stageKey, stage := range r.stages {
		logPrefix := fmt.Sprintf("[%s] ", stageKey)
		logger := log.New(os.Stdout, logPrefix, 0)
		logger.Printf("Running test: %s\n", stage.name)
		err := stage.runFunc()
		if err != nil {
			logger.Println("Test failed")
			logger.Println(err)
			return StageRunnerResult{
				failedAtStage: stage,
				error:         err,
			}
		}

		logger.Println("Test passed.")
	}

	return StageRunnerResult{
		error: nil,
	}
}

// Stage is blah
type Stage struct {
	name        string
	description string
	runFunc     func() error
}
