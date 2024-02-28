package redis_client

import (
	"bytes"
	"errors"
	"net"
	"time"

	resp "github.com/codecrafters-io/redis-tester/internal/resp"
)

type RedisClientCallbacks struct {
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
	OnValueRead func(value resp.Value)
}

type RedisClient struct {
	// Conn is the underlying connection to the Redis server.
	Conn net.Conn

	// ReadBuffer contains bytes that have been read but not decoded as a value yet.
	// It can be used to check whether there are any bytes left in the buffer after reading a value.
	ReadBuffer bytes.Buffer

	// Callbacks is a set of functions that are called at various points in the client's lifecycle.
	Callbacks RedisClientCallbacks
}

func NewRedisClient(addr string) (*RedisClient, error) {
	conn, err := newRedisConn(addr)

	if err != nil {
		return nil, err
	}

	return &RedisClient{
		Conn:       conn,
		ReadBuffer: bytes.Buffer{},
	}, nil
}

func NewRedisClientWithCallbacks(addr string, callbacks RedisClientCallbacks) (*RedisClient, error) {
	conn, err := newRedisConn(addr)

	if err != nil {
		return nil, err
	}

	return &RedisClient{
		Conn:       conn,
		ReadBuffer: bytes.Buffer{},
		Callbacks:  callbacks,
	}, nil
}

func (c *RedisClient) SendCommand(command string, args ...string) error {
	if c.Callbacks.OnSendCommand != nil {
		c.Callbacks.OnSendCommand(command, args...)
	}

	encodedValue := resp.Encode(resp.NewStringArrayValue(append([]string{command}, args...)))
	return c.SendRaw(encodedValue)
}

func (c *RedisClient) SendRaw(bytes []byte) error {
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

func (c *RedisClient) ReadValue() (resp.Value, error) {
	return c.ReadValueWithTimeout(2 * time.Second)
}

func (c *RedisClient) ReadValueWithTimeout(timeout time.Duration) (resp.Value, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			break
		}

		// Ensure we allow enough time for a read to complete
		c.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		buf := make([]byte, 1024)

		// fmt.Println("Reading")
		n, err := c.Conn.Read(buf)
		if err != nil {
			continue // Let's swallow these errors and try reading again anyway
		}

		c.ReadBuffer.Write(buf[:n])

		// Let's try to decode the value at this point.
		_, _, err = resp.Decode(c.ReadBuffer.Bytes())

		if err == nil {
			break // We were able to read a value!
		}

		if _, ok := err.(resp.InvalidRESPError); ok {
			break // We've read an invalid value, we can stop reading immediately
		}
	}

	value, readBytesCount, err := resp.Decode(c.ReadBuffer.Bytes())
	if err != nil {
		return resp.Value{}, err
	}

	// We've read a value! Let's remove the bytes we've read from the buffer
	c.ReadBuffer = *bytes.NewBuffer(c.ReadBuffer.Bytes()[readBytesCount:])
	return value, nil
}

func newRedisConn(address string) (net.Conn, error) {
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
