package linewriter

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWritesLine(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	lw := New(w, 100*time.Millisecond)
	n, _ := lw.Write([]byte{'a', 'b', 'c', '\n'})
	if !assert.Equal(t, 4, n) {
		t.FailNow()
	}

	err := lw.Flush()
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, "abc\n", w.String()) {
		t.FailNow()
	}
}

func TestWritesNewLine(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	lw := New(w, 100*time.Millisecond)
	n, _ := lw.Write([]byte{'a', 'b', 'c', '\n', '\n'})
	if !assert.Equal(t, 5, n) {
		t.FailNow()
	}

	err := lw.Flush()
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, "abc\n\n", w.String()) {
		t.FailNow()
	}
}

func TestHandlesTimeout(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	lw := New(w, 100*time.Millisecond)
	n, _ := lw.Write([]byte{'a', 'b', 'c'})
	if !assert.Equal(t, 3, n) {
		t.FailNow()
	}

	if !assert.Equal(t, "", w.String()) {
		t.FailNow()
	}

	time.Sleep(200 * time.Millisecond)

	if !assert.Equal(t, "abc\n", w.String()) {
		t.FailNow()
	}

	lw.Write([]byte{'a'})

	if !assert.Equal(t, "abc\n", w.String()) {
		t.FailNow()
	}

	time.Sleep(200 * time.Millisecond)

	if !assert.Equal(t, "abc\na\n", w.String()) {
		t.FailNow()
	}
}
