package internal

import (
	"io/ioutil"
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
	m.mockedStdout, _ = ioutil.TempFile("", "")
	m.mockedStdin, _ = ioutil.TempFile("", "")
	m.mockedStderr, _ = ioutil.TempFile("", "")

	os.Stdout = m.mockedStdout
	os.Stdin = m.mockedStdin
	os.Stderr = m.mockedStderr
}

func (m *IOMocker) ReadStdout() []byte {
	bytes, err := ioutil.ReadFile(m.mockedStdout.Name())
	if err != nil {
		panic(err)
	}

	return bytes
}

func (m *IOMocker) ReadStderr() []byte {
	bytes, err := ioutil.ReadFile(m.mockedStderr.Name())
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
