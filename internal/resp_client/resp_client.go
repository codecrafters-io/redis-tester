package resp_client

import (
	"bytes"
	"errors"
	"net"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp/decoder"
	"github.com/codecrafters-io/redis-tester/internal/resp/encoder"
)

type RespClientCallbacks struct {
	// OnSendCommand is called when a command is sent to the server.
	// This can be useful for info logs.
	OnSendCommand func(command string, args ...string)

	// OnRawSend is called when raw bytes are sent to the server.
	// This can be useful for debug logs.
	OnRawSend func(bytes []byte)

	// OnRawRead is called when raw bytes are read from the server.
	// This can be useful for debug logs.
	OnRawRead func(bytes []byte)

	// OnValueRead is called when a RESP value is decoded from bytes read from the server.
	// This can be useful for success logs.
	OnValueRead func(value resp_value.Value)
}

type RespClient struct {
	// Conn is the underlying connection to the Redis server.
	Conn net.Conn

	// UnreadBuffer contains bytes that have been read but not decoded as a value yet.
	// It can be used to check whether there are any bytes left in the buffer after reading a value.
	UnreadBuffer bytes.Buffer

	// LastValueBytes contains the bytes of the last value that was decoded.
	LastValueBytes []byte

	// Callbacks is a set of functions that are called at various points in the client's lifecycle.
	Callbacks RespClientCallbacks
}

func NewRespClient(addr string) (*RespClient, error) {
	conn, err := newConn(addr)

	if err != nil {
		return nil, err
	}

	return &RespClient{
		Conn:         conn,
		UnreadBuffer: bytes.Buffer{},
	}, nil
}

func NewRespClientWithCallbacks(addr string, callbacks RespClientCallbacks) (*RespClient, error) {
	conn, err := newConn(addr)

	if err != nil {
		return nil, err
	}

	return &RespClient{
		Conn:         conn,
		UnreadBuffer: bytes.Buffer{},
		Callbacks:    callbacks,
	}, nil
}

func (c *RespClient) Close() error {
	return c.Conn.Close()
}

func (c *RespClient) SendCommand(command string, args ...string) error {
	if c.Callbacks.OnSendCommand != nil {
		c.Callbacks.OnSendCommand(command, args...)
	}

	encodedValue := resp_encoder.Encode(resp_value.NewStringArrayValue(append([]string{command}, args...)))
	return c.SendRaw(encodedValue)
}

func (c *RespClient) SendRaw(bytes []byte) error {
	if c.Callbacks.OnRawSend != nil {
		c.Callbacks.OnRawSend(bytes)
	}

	n, err := c.Conn.Write(bytes)
	if err != nil {
		return err
	}

	// TODO: Check when this happens - is it a valid error?
	if n != len(bytes) {
		return errors.New("failed to write entire bytes to connection")
	}

	return nil
}

func (c *RespClient) ReadValue() (resp_value.Value, error) {
	return c.ReadValueWithTimeout(2 * time.Second)
}

func (c *RespClient) ReadIntoBuffer() error {
	// Ensure we allow enough time for a read to complete
	c.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	buf := make([]byte, 1024)
	n, err := c.Conn.Read(buf)

	if n > 0 {
		c.UnreadBuffer.Write(buf[:n])
	}

	return err
}

func (c *RespClient) ReadValueWithTimeout(timeout time.Duration) (resp_value.Value, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			break
		}

		// We'll swallow these errors and try reading again anyway
		_ = c.ReadIntoBuffer()

		// Let's try to decode the value at this point.
		_, _, err := resp_decoder.Decode(c.UnreadBuffer.Bytes())

		if err == nil {
			break // We were able to read a value!
		}

		if _, ok := err.(resp_decoder.InvalidRESPError); ok {
			break // We've read an invalid value, we can stop reading immediately
		}
	}

	value, readBytesCount, err := resp_decoder.Decode(c.UnreadBuffer.Bytes())
	if err != nil {
		if c.Callbacks.OnRawRead != nil {
			c.Callbacks.OnRawRead(c.UnreadBuffer.Bytes())
		}

		return resp_value.Value{}, err
	}

	// We've read a value! Let's remove the bytes we've read from the buffer
	c.LastValueBytes = c.UnreadBuffer.Bytes()[:readBytesCount]
	c.UnreadBuffer = *bytes.NewBuffer(c.UnreadBuffer.Bytes()[readBytesCount:])

	if c.Callbacks.OnValueRead != nil {
		c.Callbacks.OnValueRead(value)
	}

	return value, nil
}

func newConn(address string) (net.Conn, error) {
	attempts := 0

	for {
		var err error
		var conn net.Conn

		conn, err = net.Dial("tcp", address)

		if err == nil {
			return conn, nil
		}

		// Already a timeout
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, err
		}

		// 50 * 100ms = 5s
		if attempts > 50 {
			return nil, err
		}

		attempts += 1
		time.Sleep(100 * time.Millisecond)
	}
}
