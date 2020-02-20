package linewriter

import (
	"bytes"
	"io"
	"time"
)

type LineWriter struct {
	buffer           chan byte
	flushableStrings chan string
	writer           io.Writer
	timeout          time.Duration
	lastErr          error
	flushChan        chan bool
	flushedChan      chan bool
}

// Write queues a string for writing
func (w *LineWriter) Write(bytes []byte) (n int, err error) {
	for _, byte := range bytes {
		w.buffer <- byte
	}

	return len(bytes), nil
}

// Flush flushes any pending strings, and returns an error if any writes failed
// in the past
func (w *LineWriter) Flush() (err error) {
	w.flushChan <- true
	<-w.flushedChan
	return w.lastErr
}

// New returns a LineWriter instance
func New(w io.Writer, timeout time.Duration) *LineWriter {
	lw := &LineWriter{
		buffer:           make(chan byte),
		flushableStrings: make(chan string),
		writer:           w,
		timeout:          timeout,
		flushChan:        make(chan bool),
		flushedChan:      make(chan bool),
	}
	lw.startReader()
	lw.startWriter()
	return lw
}

func (w *LineWriter) startReader() {
	go func(w *LineWriter) {
		accumulated := bytes.NewBuffer([]byte{})

	loop:
		for {
			select {
			case b := <-w.buffer:
				accumulated.WriteByte(b)

				if string(b) == "\n" {
					w.flushableStrings <- accumulated.String()
					accumulated = bytes.NewBuffer([]byte{})
				}
			case <-time.After(w.timeout):
				if accumulated.Len() > 0 {
					accumulated.WriteByte('\n')
					w.flushableStrings <- accumulated.String()
					accumulated = bytes.NewBuffer([]byte{})
				}
			case <-w.flushChan:
				if accumulated.Len() > 0 {
					accumulated.WriteByte('\n')
					w.flushableStrings <- accumulated.String()
					accumulated = bytes.NewBuffer([]byte{})
				}
				close(w.flushableStrings)
				break loop
			}
		}
	}(w)
}

func (w *LineWriter) startWriter() {
	go func(w *LineWriter) {
	loop:
		for {
			select {
			case str, more := <-w.flushableStrings:
				if str != "" {
					_, err := w.writer.Write([]byte(str))
					if err != nil {
						w.lastErr = err
					}
				}

				if !more {
					w.flushedChan <- true
					break loop
				}
			}
		}
	}(w)
}
