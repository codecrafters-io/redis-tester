package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// RESP types
const (
	RESP_SIMPLE_STRING = '+'
	RESP_ERROR         = '-'
	RESP_INTEGER       = ':'
	RESP_BULK_STRING   = '$'
	RESP_ARRAY         = '*'
)

type RESPCodec struct{}

func NewRESPCodec() *RESPCodec {
	return &RESPCodec{}
}

func (r *RESPCodec) ReadCommand(reader *bufio.Reader) ([]string, error) {
	// first byte determines the RESP type
	firstByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if firstByte != RESP_ARRAY {
		return nil, fmt.Errorf("expected array, got %c", firstByte)
	}

	// array length
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", lengthStr)
	}

	// array elements
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

	// string length
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %s", lengthStr)
	}

	// string content
	content := make([]byte, length)
	_, err = reader.Read(content)
	if err != nil {
		return "", err
	}

	// trailing \r\n
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (r *RESPCodec) EncodeSimpleString(str string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", str))
}

func (r *RESPCodec) EncodeBulkString(str string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

func (r *RESPCodec) EncodeNil() []byte {
	return []byte("$-1\r\n")
}

func (r *RESPCodec) EncodeError(err string) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err))
}

func (r *RESPCodec) EncodeInteger(i int) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", i))
}

func (r *RESPCodec) EncodeArray(elements []string) []byte {
	if len(elements) == 0 {
		return []byte("*0\r\n")
	}

	var result []byte
	result = append(result, []byte(fmt.Sprintf("*%d\r\n", len(elements)))...)

	for _, element := range elements {
		result = append(result, r.EncodeBulkString(element)...)
	}

	return result
}
