package resp_connection

import (
	"bytes"
	"errors"
	"net"
	"time"

	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
	resp_encoder "github.com/codecrafters-io/redis-tester/internal/resp/encoder"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RespConnectionCallbacks struct {
	// BeforeSendCommand is called when a command is sent to the server.
	// This can be useful for info logs.
	BeforeSendCommand func(reusedConnection bool, command string, args ...string)

	// BeforeSendValue is called when a value is sent using SendValue.
	// It is NOT called when using SendCommand
	BeforeSendValue func(response resp_value.Value)

	// BeforeSendBytes is called when raw bytes are sent to the server.
	// This can be useful for debug logs.
	BeforeSendBytes func(bytes []byte)

	// AfterBytesReceived is called when raw bytes are read from the server.
	// This can be useful for debug logs.
	AfterBytesReceived func(bytes []byte)

	// AfterReadValue is called when a RESP value is decoded from bytes read from the server.
	// This can be useful for success logs.
	AfterReadValue func(value resp_value.Value)
}

type RespConnection struct {
	// Callbacks is a set of functions that are called at various points in the connection's lifecycle.
	Callbacks RespConnectionCallbacks

	// Conn is the underlying connection to the Redis server.
	// TODO: Make conn private, only expose whatever clients need
	Conn net.Conn

	// LastValueBytesCount is the number of bytes in which the last value was decoded
	// It is needed to send REPLCONF ACK <count>
	// Where <count> is the number of bytes received from the master except for the count of bytes
	// of last decoded value
	LastValueBytesCount int

	// ReadValueInterceptor is a function that is called when a value is read from the connection.
	// It can be used to intercept a read value and return a different value.
	ReadValueInterceptor func(value resp_value.Value) resp_value.Value

	// ReceivedBytesCount is the number of bytes received using this connection.
	ReceivedBytesCount int

	// Offset tracking:
	// SentBytesCount is the number of bytes sent using this connection, but this can be mutated when it makes sense (ex: handshake bytes are not counted by redis for acks)
	SentBytesCount int

	// TotalSentBytes is the total number of bytes sent using this connection, and is not reset or changed when the connection is reset
	// This should be used for deciding if connection is new / reused and commands be logged as such
	TotalSentBytesCount int

	// UnreadBuffer contains bytes that have been read but not decoded as a value yet.
	// It can be used to check whether there are any bytes left in the buffer after reading a value.
	// TODO: Make this private, only expose buffer length or whatever is needed
	UnreadBuffer bytes.Buffer
}

func NewRespConnectionFromAddr(addr string, callbacks RespConnectionCallbacks) (*RespConnection, error) {
	conn, err := newConn(addr)

	if err != nil {
		return nil, err
	}

	return &RespConnection{
		Conn:         conn,
		UnreadBuffer: bytes.Buffer{},
		Callbacks:    callbacks,
	}, nil
}

func NewRespConnectionFromConn(conn net.Conn, callbacks RespConnectionCallbacks) (*RespConnection, error) {
	return &RespConnection{
		Conn:         conn,
		UnreadBuffer: bytes.Buffer{},
		Callbacks:    callbacks,
	}, nil
}

func (c *RespConnection) Close() error {
	return c.Conn.Close()
}

func (c *RespConnection) SendCommand(command string, args ...string) error {
	if c.Callbacks.BeforeSendCommand != nil {
		if c.TotalSentBytesCount > 0 {
			c.Callbacks.BeforeSendCommand(true, command, args...)
		} else {
			c.Callbacks.BeforeSendCommand(false, command, args...)
		}
	}

	encodedValue := resp_encoder.Encode(resp_value.NewStringArrayValue(append([]string{command}, args...)))
	return c.SendBytes(encodedValue)
}

func (c *RespConnection) SendValue(response resp_value.Value) error {
	if c.Callbacks.BeforeSendValue != nil {
		c.Callbacks.BeforeSendValue(response)
	}

	encodedValue := resp_encoder.Encode(response)
	return c.SendBytes(encodedValue)
}

func (c *RespConnection) SendBytes(bytes []byte) error {
	if c.Callbacks.BeforeSendBytes != nil {
		c.Callbacks.BeforeSendBytes(bytes)
	}

	n, err := c.Conn.Write(bytes)
	if err != nil {
		return err
	}

	c.SentBytesCount += len(bytes)
	c.TotalSentBytesCount += len(bytes)

	// TODO: Check when this happens - is it a valid error?
	if n != len(bytes) {
		return errors.New("failed to write entire bytes to connection")
	}

	return nil
}

func (c *RespConnection) ReadFullResyncRDBFile() ([]byte, error) {
	shouldStopReadingIntoBuffer := func(buf []byte) bool {
		_, _, err := resp_decoder.DecodeFullResyncRDBFile(c.UnreadBuffer.Bytes())

		if err == nil {
			return true // We were able to read a value!
		}

		if _, ok := err.(resp_decoder.InvalidInputError); ok {
			return true // We've read an invalid value, we can stop reading immediately
		}

		return false
	}

	c.readIntoBufferUntil(shouldStopReadingIntoBuffer, 2*time.Second)

	value, decodedBytesCount, err := resp_decoder.DecodeFullResyncRDBFile(c.UnreadBuffer.Bytes())

	readBytesCount := decodedBytesCount
	if err != nil {
		readBytesCount = c.UnreadBuffer.Len()
	}

	readBytes := c.UnreadBuffer.Bytes()[:readBytesCount]

	if c.Callbacks.AfterBytesReceived != nil && readBytesCount > 0 {
		c.Callbacks.AfterBytesReceived(readBytes)
	}

	if err != nil {
		return nil, err
	}

	// We've read a value! Let's remove the bytes we've read from the buffer
	c.UnreadBuffer = *bytes.NewBuffer(c.UnreadBuffer.Bytes()[readBytesCount:])
	c.ReceivedBytesCount += readBytesCount
	c.LastValueBytesCount = readBytesCount

	return value, nil
}

func (c *RespConnection) ReadValue() (resp_value.Value, error) {
	return c.ReadValueWithTimeout(2 * time.Second)
}

func (c *RespConnection) ReadIntoBuffer() error {
	// We don't want to block indefinitely, so we'll set a read deadline
	c.Conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))

	buf := make([]byte, 1024)
	n, err := c.Conn.Read(buf)

	if n > 0 {
		c.UnreadBuffer.Write(buf[:n])
	}

	return err
}

func (c *RespConnection) ReadValueWithTimeout(timeout time.Duration) (resp_value.Value, error) {
	c.readIntoBufferUntil(isRespStreamCompleteOrInvalid, timeout)

	value, decodedBytesCount, err := resp_decoder.Decode(c.UnreadBuffer.Bytes())
	if err != nil {
		if c.Callbacks.AfterBytesReceived != nil && c.UnreadBuffer.Len() > 0 {
			c.Callbacks.AfterBytesReceived(c.UnreadBuffer.Bytes())
		}

		return resp_value.Value{}, err
	}

	// We've read a value
	valueBytes := c.UnreadBuffer.Bytes()[:decodedBytesCount]
	c.ReceivedBytesCount += decodedBytesCount
	c.UnreadBuffer = *bytes.NewBuffer(c.UnreadBuffer.Bytes()[decodedBytesCount:])
	c.LastValueBytesCount = decodedBytesCount

	if c.ReadValueInterceptor != nil {
		value = c.ReadValueInterceptor(value)
		valueBytes = resp_encoder.Encode(value)
	}

	if c.Callbacks.AfterBytesReceived != nil {
		c.Callbacks.AfterBytesReceived(valueBytes)
	}

	if c.Callbacks.AfterReadValue != nil {
		c.Callbacks.AfterReadValue(value)
	}

	return value, nil
}

func (c *RespConnection) readIntoBufferUntil(condition func([]byte) bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			break
		}

		// We'll swallow these errors and try reading again anyway
		_ = c.ReadIntoBuffer()

		if condition(c.UnreadBuffer.Bytes()) {
			break
		} else {
			time.Sleep(10 * time.Millisecond) // Let's wait a bit before trying again
		}
	}
}

func (c *RespConnection) ResetByteCounters() {
	c.ReceivedBytesCount = 0
	c.SentBytesCount = 0
}

func (c *RespConnection) UpdateCallBacks(newCallBacks RespConnectionCallbacks) {
	c.Callbacks = newCallBacks
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

func isRespStreamCompleteOrInvalid(buf []byte) bool {
	_, _, err := resp_decoder.Decode(buf)

	if err == nil {
		return true // Complete value
	}

	if _, ok := err.(resp_decoder.InvalidInputError); ok {
		return true // We've read an invalid value, we can stop reading immediately
	}

	return false
}
