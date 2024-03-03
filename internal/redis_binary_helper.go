package internal

import (
	"strings"

	executable "github.com/codecrafters-io/tester-utils/executable"
	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type RedisBinary struct {
	executable *executable.Executable
	logger     *logger.Logger
	args       []string
}

func NewRedisBinary(stageHarness *test_case_harness.TestCaseHarness) *RedisBinary {
	b := &RedisBinary{
		executable: stageHarness.NewExecutable(),
		logger:     stageHarness.Logger,
	}

	stageHarness.RegisterTeardownFunc(func() { b.Kill() })

	return b
}

func (b *RedisBinary) Run() error {
	if b.args == nil || len(b.args) == 0 {
		b.logger.Infof("$ ./spawn_redis_server.sh")
	} else {
		b.logger.Infof("$ ./spawn_redis_server.sh %s", strings.Join(b.args, " "))
	}

	if err := b.executable.Start(b.args...); err != nil {
		return err
	}

	return nil
}

func (b *RedisBinary) HasExited() bool {
	return b.executable.HasExited()
}

func (b *RedisBinary) Kill() error {
	b.logger.Debugf("Terminating program")
	if err := b.executable.Kill(); err != nil {
		b.logger.Debugf("Error terminating program: '%v'", err)
		return err
	}

	b.logger.Debugf("Program terminated successfully")
	return nil // When does this happen?
}