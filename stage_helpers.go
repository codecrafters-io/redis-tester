package main

import "time"

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

	// Wait for Redis program to boot
	time.Sleep(2 * time.Second)

	return nil
}

func (b *RedisBinary) Kill() error {
	b.logger.Debugf("Terminating program")
	b.executable.Kill()
	b.logger.Debugf("Program terminated successfully")
	return nil // When does this happen?
}
