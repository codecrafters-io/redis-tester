package internal

import (
	"context"
	"net"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	executable "github.com/codecrafters-io/tester-utils/executable"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

type RedisBinary struct {
	executable *executable.Executable
	logger     *logger.Logger
	args       []string
}

func NewRedisBinary(stageHarness *testerutils.StageHarness) *RedisBinary {
	b := &RedisBinary{
		executable: stageHarness.Executable,
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

	// We don't sleep anymore, since the redis client now uses a "retry-ing" dialer
	// time.Sleep(1000 * time.Millisecond) // Redis clients perform 20 retries with a 500ms backoff, so let's keep this short-ish.

	// ctx := context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	// defer cancel()
	// b.waitForPort(ctx)

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

func (b *RedisBinary) waitForPort(ctx context.Context) {
	dialedChan := make(chan bool)
	go func(ctx context.Context, dialedChan chan<- bool) {
		for {
			_, err := net.Dial("tcp", "localhost:6379")
			if err == nil {
				dialedChan <- true
				break
			}

			select {
			case <-time.After(100 * time.Millisecond):
				continue
			case <-ctx.Done():
				break
			}
		}
	}(ctx, dialedChan)

	// Wait either until Dial works, or until the ctx times out
	select {
	case <-dialedChan:
		return
	case <-ctx.Done():
		return
	}
}
