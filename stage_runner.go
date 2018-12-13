package main

// StageRunnerResult is returned from StageRunner.Run()
type StageRunnerResult struct {
	failedAtStage int
	error         error
}

// IsSuccess says whether a StageRunnerResult was successful
// or not
func (res StageRunnerResult) IsSuccess() bool {
	return res.error == nil
}

// StageRunner is used to run multiple stages
type StageRunner struct {
	stages map[string]func
}

func newStageRunner() StageRunner {
	return StageRunner{}
}

// Run tests in a specific StageRunner
func (r StageRunner) Run() StageRunnerResult {
	return StageRunnerResult{
		failedAtStage: 0,
		error:         nil,
	}
}

type Stage struct {
}
