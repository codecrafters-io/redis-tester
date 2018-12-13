package main

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
	for _, stage := range r.stages {
		err := stage.runFunc()
		if err != nil {
			return StageRunnerResult{
				failedAtStage: stage,
				error:         err,
			}
		}

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
