package main

// StageRunnerResult is returned from StageRunner.Run()
type StageRunnerResult struct {
	failedAtStage Stage
	error         error
	logger        *customLogger
}

// IsSuccess says whether a StageRunnerResult was successful
// or not
func (res StageRunnerResult) IsSuccess() bool {
	return res.error == nil
}

// StageRunner is used to run multiple stages
type StageRunner struct {
	stages  map[string]Stage
	isDebug bool
}

func newStageRunner(isDebug bool) StageRunner {
	return StageRunner{
		isDebug: isDebug,
		stages: map[string]Stage{
			"stage-1": Stage{
				name:    "Stage 1: PING <-> PONG",
				logger:  getLogger(isDebug, "[stage-1] "),
				runFunc: runStage1,
			},
			"stage-2": Stage{
				name:    "Stage 2: Multiple Clients",
				logger:  getLogger(isDebug, "[stage-2] "),
				runFunc: runStage2,
			},
		},
	}
}

// Run tests in a specific StageRunner
func (r StageRunner) Run() StageRunnerResult {
	for _, stage := range r.stages {
		logger := stage.logger
		logger.Infof("Running test: %s", stage.name)
		err := stage.runFunc(logger)
		if err != nil {
			logger.Errorf("Test failed")
			logger.Errorf("%s", err)
			return StageRunnerResult{
				failedAtStage: stage,
				error:         err,
			}
		}

		logger.Successf("Test passed.")
	}

	return StageRunnerResult{
		error: nil,
	}
}

// Stage is blah
type Stage struct {
	name        string
	description string
	runFunc     func(logger *customLogger) error
	logger      *customLogger
}
