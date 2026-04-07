package redis_executable

import (
	"fmt"
	"path"
	"strings"

	executable "github.com/codecrafters-io/tester-utils/executable"
	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type RedisExecutable struct {
	executable *executable.Executable
	logger     *logger.Logger
	args       []string

	// isRunning is set to true after a successful .Start()
	// and set to false after a successful .Kill()
	// This is useful when killing executable mid-test and re-spawning it
	isRunning bool
}

func NewRedisExecutable(stageHarness *test_case_harness.TestCaseHarness) *RedisExecutable {
	b := &RedisExecutable{
		executable: stageHarness.NewExecutable(),
		logger:     stageHarness.Logger,
	}

	stageHarness.RegisterTeardownFunc(func() { b.Kill() })

	return b
}

func (b *RedisExecutable) Run(args ...string) error {
	b.args = args
	if len(b.args) == 0 {
		b.logger.Infof("$ ./%s", path.Base(b.executable.Path))
	} else {
		var log string
		log += fmt.Sprintf("$ ./%s", path.Base(b.executable.Path))
		for _, arg := range b.args {
			if strings.Contains(arg, " ") {
				log += " \"" + arg + "\""
			} else {
				log += " " + arg
			}
		}
		b.logger.Infof("%s", log)
	}

	if err := b.executable.Start(b.args...); err != nil {
		return err
	}

	b.isRunning = true
	return nil
}

func (b *RedisExecutable) HasExited() bool {
	return b.executable.HasExited()
}

func (b *RedisExecutable) Kill() error {
	if !b.isRunning {
		return nil
	}

	b.logger.Debugf("Terminating program")
	if err := b.executable.Kill(); err != nil {
		b.logger.Debugf("Error terminating program: '%v'", err)
		return err
	}

	b.logger.Debugf("Program terminated successfully")

	b.isRunning = false
	return nil // When does this happen?
}
