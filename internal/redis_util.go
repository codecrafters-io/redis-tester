package internal

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/smallnest/resp3"
)

func encodeInteger(v Value) ([]byte, error) {
	intValue, err := strconv.Atoi(string(v.data))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(":%d\r\n", intValue)), nil
}

func encodeSimpleString(v Value) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", v.data))
}

func encodeBulkString(v Value) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.data), v.data))
}

func encodeError(v Value) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", v.data))
}

func encodeArray(v Value) ([]byte, error) {
	res := []byte{}
	for _, elem := range v.array {
		val, err := elem.Encode()
		if err != nil {
			return nil, err
		}
		res = append(res, val...)
	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(v.array), res)), nil
}

func readToken(byteStream *resp3.Reader) ([]byte, error) {
	bytes, err := byteStream.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// discard \r\n
	return bytes[:len(bytes)-2], nil
}

func decodeSimpleString(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, err
	}
	return NewSimpleStringValue(string(t)), nil
}

func decodeInteger(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, nil
	}

	num, err := strconv.Atoi(string(t))
	if err != nil {
		return Value{}, err
	}
	return NewIntegerValue(num), nil
}

func decodeError(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, err
	}
	return NewErrorValue(string(t)), nil
}

func decodeBulkString(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, nil
	}

	size, err := strconv.Atoi(string(t))
	if err != nil {
		return Value{}, err
	}
	str := make([]byte, size+2)
	_, err = io.ReadFull(byteStream, str)
	if err != nil {
		return Value{}, err
	}

	// Assert \r\n over here, before discarding \r\n
	if string(str[size:]) != "\r\n" {
		return Value{}, fmt.Errorf("Expected CRLF at the end.")
	}
	str = str[:size]

	return NewBulkStringValue(string(str)), nil
}

func decodeBulkStringRDB(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, nil
	}

	size, err := strconv.Atoi(string(t))
	if err != nil {
		return Value{}, err
	}

	// RDB files when sent, don't end with \r\n, we need to reduce
	// size for reading them.
	str := make([]byte, size)

	_, err = io.ReadFull(byteStream, str)
	if err != nil {
		return Value{}, err
	}

	if byteStream.Buffered() > 0 {
		return Value{}, fmt.Errorf("Unexpected CRLF at the end.")
	}
	str = str[:size]

	return NewBulkStringValue(string(str)), nil
}
func decodeArray(byteStream *resp3.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, nil
	}
	length, err := strconv.Atoi(string(t))
	if err != nil {
		return Value{}, err
	}

	arr := make([]Value, length)
	for i := 0; i < len(arr); i++ {
		v, err := Decode(byteStream)
		if err != nil {
			return Value{}, err
		}
		arr[i] = v
	}

	return NewArrayValue(arr), nil
}

func parseRESPCommandRDB(reader *resp3.Reader) (Value, error) {
	req, err := DecodeRDB(reader)
	if err != nil {
		return Value{}, err
	}
	return req, nil
}

func SendError(err error) []byte {
	e := NewErrorValue("ERR - " + err.Error())
	return encodeError(e)
}

func SendNil() []byte {
	return []byte("$-1\r\n")
}

func SendRDBFile() []byte {
	hexStr := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Printf("Encountered %s while deconding hex string", err.Error())
		return SendError(err)
	}
	resp := []byte("$" + strconv.Itoa(len(bytes)) + "\r\n")
	return (append(resp, bytes...))
}
