package redis_client

import (
	"errors"
	"net"
	"time"

	resp "github.com/codecrafters-io/redis-tester/internal/resp"
)

type RedisClient struct {
	Conn       net.Conn
	ReadBuffer []byte
}

func NewRedisClient(addr string) (*RedisClient, error) {
	conn, err := newRedisConn(addr)

	if err != nil {
		return nil, err
	}

	return &RedisClient{
		Conn:       conn,
		ReadBuffer: make([]byte, 1024),
	}, nil
}

func (c *RedisClient) SendCommand(command string, args ...string) error {
	encodedValue := resp.Encode(resp.NewStringArrayValue(append([]string{command}, args...)))

	n, err := c.Conn.Write(encodedValue)
	if err != nil {
		return err
	}

	// TODO: Check when this happens - is it a valid error?
	if n != len(encodedValue) {
		return errors.New("failed to write entire command to connection")
	}

	return nil
}

func (c *RedisClient) ReadMessage() (resp.Value, error) {
	return c.ReadMessageWithTimeout(2 * time.Second)
}

func (c *RedisClient) ReadMessageWithTimeout(timeout time.Duration) (resp.Value, error) {
	c.Conn.SetReadDeadline(time.Now().Add(timeout))

	n, err := c.Conn.Read(c.ReadBuffer)
	if err != nil {
		return resp.Value{}, err
	}

	value, _, err := resp.Decode(c.ReadBuffer[:n])
	if err != nil {
		return resp.Value{}, err
	}

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
