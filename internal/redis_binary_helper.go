package internal

import (
	"context"
	"net"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
)

type RedisBinary struct {
	executable *testerutils.Executable
	logger     *testerutils.Logger
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
	b.logger.Debugf("Running program")
	if err := b.executable.Start(); err != nil {
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
