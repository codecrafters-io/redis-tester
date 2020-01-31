package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// Executable represents a program that can be executed
type Executable struct {
	path           string
	timeoutInSecs  int
	suppressOutput bool

	// These are set & removed together
	cmd          *exec.Cmd
	stdoutPipe   io.ReadCloser
	stderrPipe   io.ReadCloser
	stdoutBytes  []byte
	stderrBytes  []byte
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer

	stdoutEchoed chan (bool)
	stderrEchoed chan (bool)
}

// ExecutableResult holds the result of an executable run
type ExecutableResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// NewExecutable returns an Executable struct
func NewExecutable(path string) *Executable {
	return &Executable{path: path, timeoutInSecs: 10, suppressOutput: true}
}

func NewVerboseExecutable(path string) *Executable {
	return &Executable{path: path, timeoutInSecs: 10, suppressOutput: false}
}

func (e *Executable) isRunning() bool {
	return e.cmd != nil
}

// Start starts the specified command but does not wait for it to complete.
func (e *Executable) Start(args ...string) error {
	var err error

	if e.isRunning() {
		return errors.New("process already in progress")
	}

	// TODO: Use timeout!
	e.cmd = exec.Command(e.path, args...)

	// Setup stdout capture
	e.stdoutPipe, err = e.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	e.stdoutBytes = []byte{}
	e.stdoutBuffer = bytes.NewBuffer(e.stdoutBytes)
	e.setupIORelay(e.stdoutPipe, os.Stdout, e.stdoutBuffer)

	// Setup stderr relay
	e.stderrPipe, err = e.cmd.StderrPipe()
	if err != nil {
		return err
	}
	e.stderrBytes = []byte{}
	e.stderrBuffer = bytes.NewBuffer(e.stderrBytes)
	e.setupIORelay(e.stderrPipe, os.Stderr, e.stderrBuffer)

	return e.cmd.Start()
}

func (e *Executable) setupIORelay(childReader io.Reader, parentWriter io.Writer, buffer io.Writer) {
	go func() {
		writer := buffer
		if !e.suppressOutput {
			writer = io.MultiWriter(parentWriter, writer)
		}
		ioutil.ReadAll(io.TeeReader(childReader, writer))
	}()

}

// Run starts the specified command, waits for it to complete and returns the
// result.
func (e *Executable) Run(args ...string) (ExecutableResult, error) {
	var err error

	if err = e.Start(args...); err != nil {
		return ExecutableResult{}, err
	}

	return e.Wait()
}

// Wait waits for the program to finish and results the result
func (e *Executable) Wait() (ExecutableResult, error) {
	e.cmd.Wait()

	stdout := e.stdoutBuffer.Bytes()
	stderr := e.stderrBuffer.Bytes()

	defer func() {
		e.cmd = nil
		e.stdoutPipe = nil
		e.stderrPipe = nil
		e.stdoutBuffer = nil
		e.stderrBuffer = nil
		e.stdoutBytes = nil
		e.stderrBytes = nil
	}()

	return ExecutableResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: e.cmd.ProcessState.ExitCode(),
	}, nil
}
