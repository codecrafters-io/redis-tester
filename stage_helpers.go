package main

import (
	"context"
	"net"
	"time"
)

type RedisBinary struct {
	executable *Executable
	logger     *customLogger
}

func NewRedisBinary(executable *Executable, logger *customLogger) *RedisBinary {
	return &RedisBinary{
		executable: executable,
		logger:     logger,
	}
}

func (b *RedisBinary) Run() error {
	b.logger.Debugf("Running program")
	if err := b.executable.Start(); err != nil {
		return err
	}

	time.Sleep(2000 * time.Millisecond)

	// ctx := context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	// defer cancel()
	// b.waitForPort(ctx)

	return nil
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
