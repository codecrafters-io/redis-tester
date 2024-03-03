package internal

import (
	"os"
)

type IOMocker struct {
	originalStdout *os.File
	originalStderr *os.File
	originalStdin  *os.File
	mockedStdout   *os.File
	mockedStderr   *os.File
	mockedStdin    *os.File
}

func NewStdIOMocker() *IOMocker {
	return &IOMocker{}
}

func (m *IOMocker) Start() {
	m.originalStdout = os.Stdout
	m.originalStdin = os.Stdin
	m.originalStdin = os.Stdin

	m.Reset()
}

func (m *IOMocker) Reset() {
	m.mockedStdout, _ = os.CreateTemp("", "")
	m.mockedStdin, _ = os.CreateTemp("", "")
	m.mockedStderr, _ = os.CreateTemp("", "")

	os.Stdout = m.mockedStdout
	os.Stdin = m.mockedStdin
	os.Stderr = m.mockedStderr
}

func (m *IOMocker) ReadStdout() []byte {
	bytes, err := os.ReadFile(m.mockedStdout.Name())
	if err != nil {
		panic(err)
	}

	return bytes
}

func (m *IOMocker) ReadStderr() []byte {
	bytes, err := os.ReadFile(m.mockedStderr.Name())
	if err != nil {
		panic(err)
	}

	return bytes
}

func (m *IOMocker) End() {
	os.Stdout = m.originalStdout
	os.Stdin = m.originalStdin
	os.Stderr = m.originalStderr
}
