package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

const (
	RESP_SIMPLE_STRING = '+'
	RESP_ERROR         = '-'
	RESP_BULK_STRING   = '$'
	RESP_ARRAY         = '*'
)

type RESPCodec struct{}

func NewRESPCodec() *RESPCodec {
	return &RESPCodec{}
}

func (r *RESPCodec) ReadCommand(reader *bufio.Reader) ([]string, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if firstByte != RESP_ARRAY {
		return nil, fmt.Errorf("expected array, got %c", firstByte)
	}

	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", lengthStr)
	}

	args := make([]string, length)
	for i := 0; i < length; i++ {
		arg, err := r.readBulkString(reader)
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}

	return args, nil
}

func (r *RESPCodec) readBulkString(reader *bufio.Reader) (string, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	if firstByte != RESP_BULK_STRING {
		return "", fmt.Errorf("expected bulk string, got %c", firstByte)
	}

	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %s", lengthStr)
	}

	content := make([]byte, length)
	_, err = reader.Read(content)
	if err != nil {
		return "", err
	}

	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (r *RESPCodec) EncodeSimpleString(str string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", str))
}

func (r *RESPCodec) EncodeError(err string) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err))
}
