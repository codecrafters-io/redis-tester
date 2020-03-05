package main

import (
	"fmt"

	"math/rand"
	"time"
)

// StageRunnerResult is returned from StageRunner.Run()
type StageRunnerResult struct {
	lastStageIndex int
	error          error
	logger         *customLogger
}

// IsSuccess says whether a StageRunnerResult was successful
// or not
func (res StageRunnerResult) IsSuccess() bool {
	return res.error == nil
}

// StageRunner is used to run multiple stages
type StageRunner struct {
	stages  []Stage
	isDebug bool
}

func newStageRunner(isDebug bool) StageRunner {
	return StageRunner{
		isDebug: isDebug,
		stages: []Stage{
			Stage{
				slug:    "init",
				name:    "Stage 1: Bind to a port",
				logger:  getLogger(isDebug, "[stage-1] "),
				runFunc: testBindToPort,
			},
			Stage{
				slug:    "ping-pong",
				name:    "Stage 2: Respond to PING",
				logger:  getLogger(isDebug, "[stage-2] "),
				runFunc: testPingPongOnce,
			},
			Stage{
				slug:    "ping-pong-multiple",
				name:    "Stage 3: Respond to multiple PINGs",
				logger:  getLogger(isDebug, "[stage-3] "),
				runFunc: testPingPongMultiple,
			},
			Stage{
				slug:    "concurrent-clients",
				name:    "Stage 4: Handle concurrent clients",
				logger:  getLogger(isDebug, "[stage-4] "),
				runFunc: testPingPongConcurrent,
			},
			Stage{
				slug:    "echo",
				name:    "Stage 5: Implement the ECHO command",
				logger:  getLogger(isDebug, "[stage-5] "),
				runFunc: testEcho,
			},
			Stage{
				slug:    "set_get",
				name:    "Stage 6: SET & GET",
				logger:  getLogger(isDebug, "[stage-6] "),
				runFunc: testGetSet,
			},
			Stage{
				slug:    "expiry",
				name:    "Stage 7: Expiry!",
				logger:  getLogger(isDebug, "[stage-7] "),
				runFunc: testExpiry,
			},
		},
	}
}

// Run tests in a specific StageRunner
func (r StageRunner) Run(executable *Executable) StageRunnerResult {
	for index, stage := range r.stages {
		logger := stage.logger
		logger.Infof("Running test: %s", stage.name)

		stageResultChannel := make(chan error, 1)
		go func() {
			err := stage.runFunc(executable, logger)
			stageResultChannel <- err
		}()

		var err error
		select {
		case stageErr := <-stageResultChannel:
			err = stageErr
		case <-time.After(10 * time.Second):
			err = fmt.Errorf("timed out, test exceeded 10 seconds")
		}

		if err != nil {
			reportTestError(err, r.isDebug, logger)
			return StageRunnerResult{
				lastStageIndex: index,
				error:          err,
			}
		}

		logger.Successf("Test passed.")
	}

	return StageRunnerResult{
		lastStageIndex: len(r.stages) - 1,
		error:          nil,
	}
}

func (r StageRunner) StageCount() int {
	return len(r.stages)
}

// Truncated returns a stageRunner with fewer stages
func (r StageRunner) Truncated(stageSlug string) StageRunner {
	newStages := make([]Stage, 0)
	for _, stage := range r.stages {
		newStages = append(newStages, stage)
		if stage.slug == stageSlug {
			return StageRunner{
				isDebug: r.isDebug,
				stages:  newStages,
			}
		}
	}

	panic(fmt.Sprintf("Stage slug %v not found. Stages: %v", stageSlug, r.stages))
}

// Fuck you, go
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Randomized returns a stage runner that has stages randomized
func (r StageRunner) Randomized() StageRunner {
	return StageRunner{
		isDebug: r.isDebug,
		stages:  shuffleStages(r.stages),
	}
}

func shuffleStages(stages []Stage) []Stage {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]Stage, len(stages))
	perm := r.Perm(len(stages))
	for i, randIndex := range perm {
		ret[i] = stages[randIndex]
	}
	return ret
}

func reportTestError(err error, isDebug bool, logger *customLogger) {
	logger.Errorf("%s", err)
	if isDebug {
		logger.Errorf("Test failed")
	} else {
		logger.Errorf("Test failed " +
			"(try setting 'debug: true' in your codecrafters.yml to see more details)")
	}
}

// Stage is blah
type Stage struct {
	slug    string
	name    string
	runFunc func(executable *Executable, logger *customLogger) error
	logger  *customLogger
}
