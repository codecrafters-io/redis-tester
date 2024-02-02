package internal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Type struct {
	slug rune
}

func (t Type) Rune() rune {
	return t.slug
}

func GetType(r rune) (Type, error) {
	switch r {
	case SIMPLE_STRING.slug:
		return SIMPLE_STRING, nil
	case BULK_STRING.slug:
		return BULK_STRING, nil
	case INTEGER.slug:
		return INTEGER, nil
	case ERROR.slug:
		return ERROR, nil
	case ARRAY.slug:
		return ARRAY, nil
	}
	return INVALID_TYPE, fmt.Errorf("unknown type: %c", r)
}

var (
	INVALID_TYPE  = Type{}
	SIMPLE_STRING = Type{'+'}
	INTEGER       = Type{':'}
	BULK_STRING   = Type{'$'}
	ARRAY         = Type{'*'}
	ERROR         = Type{'-'}
)

type Value struct {
	typ   Type
	data  []byte
	array []Value
}

func NewValue(b []byte, v []Value, t Type) (Value, error) {
	if v != nil || t == ARRAY {
		return Value{
			typ:   t,
			array: v,
		}, nil
	}
	return Value{
		typ:  t,
		data: b,
	}, nil
}

func NewSimpleStringValue(s string) Value {
	v, _ := NewValue([]byte(s), nil, SIMPLE_STRING)
	return v
}

func NewBulkStringValue(s string) Value {
	v, _ := NewValue([]byte(s), nil, BULK_STRING)
	return v
}

func NewIntegerValue(i int) Value {
	v, _ := NewValue([]byte(fmt.Sprint(i)), nil, INTEGER)
	return v
}

func NewErrorValue(err string) Value {
	v, _ := NewValue([]byte(err), nil, ERROR)
	return v
}

func NewArrayValue(arr []Value) Value {
	v, _ := NewValue(nil, arr, ARRAY)
	return v
}

func (v *Value) Type() Type {
	return v.typ
}

func (v *Value) Data() []byte {
	return v.data
}

func (v *Value) Array() []Value {
	if v.typ == ARRAY {
		return v.array
	}
	return []Value{}
}

func (v *Value) String() string {
	return string(v.data)
}

func (v *Value) Integer() (int, error) {
	return strconv.Atoi(string(v.data))
}

func (v Value) Encode() ([]byte, error) {
	switch v.typ {
	case INTEGER:
		return encodeInteger(v)
	case SIMPLE_STRING:
		return encodeSimpleString(v), nil
	case BULK_STRING:
		return encodeBulkString(v), nil
	case ERROR:
		return encodeError(v), nil
	case ARRAY:
		return encodeArray(v)
	}
	return []byte{}, fmt.Errorf("Encode was given an unsupported type")
}

func Decode(byteStream *bufio.Reader, decodeRdb bool) (Value, error) {
	b, err := byteStream.ReadByte()
	if err != nil {
		return Value{}, err
	}

	bType, err := GetType(rune(b))
	if err != nil {
		return Value{}, err
	}

	switch bType {
	case SIMPLE_STRING:
		return decodeSimpleString(byteStream)
	case BULK_STRING:
		return decodeBulkString(byteStream, decodeRdb)
	case ARRAY:
		return decodeArray(byteStream)
	case INTEGER:
		return decodeInteger(byteStream)
	case ERROR:
		return decodeError(byteStream)
	}
	return Value{}, fmt.Errorf("Decode was given an unsupported data type %c: ", bType)
}

func encodeInteger(v Value) ([]byte, error) {
	int_value, err := strconv.Atoi(string(v.data))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(":%d\r\n", int_value)), nil
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
	return []byte(fmt.Sprintf("*%d\r\n%s\r\n", len(res), res)), nil
}

func readToken(byteStream *bufio.Reader) ([]byte, error) {
	bytes, err := byteStream.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// discard \r\n
	return bytes[:len(bytes)-2], nil
}

func decodeSimpleString(byteStream *bufio.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, err
	}
	return NewSimpleStringValue(string(t)), nil
}

func decodeInteger(byteStream *bufio.Reader) (Value, error) {
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

func decodeError(byteStream *bufio.Reader) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, err
	}
	return NewErrorValue(string(t)), nil
}

func decodeBulkString(byteStream *bufio.Reader, decodeRdb bool) (Value, error) {
	t, err := readToken(byteStream)
	if err != nil {
		return Value{}, nil
	}

	size, err := strconv.Atoi(string(t))
	if err != nil {
		return Value{}, err
	}

	extend := 0 // RDB files when sent, don't end with \r\n, we need to reduce
	// size for reading them.
	if decodeRdb == true {
		extend += 0
	} else {
		extend += 2
	}
	str := make([]byte, size+extend)

	_, err = io.ReadFull(byteStream, str)
	if err != nil {
		return Value{}, err
	}

	// discard \r\n
	str = str[:size]

	return NewBulkStringValue(string(str)), nil
}

func decodeArray(byteStream *bufio.Reader) (Value, error) {
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
		v, err := Decode(byteStream, false)
		if err != nil {
			return Value{}, err
		}
		arr[i] = v
	}

	return NewArrayValue(arr), nil
}

func SendError(err error) []byte {
	e := NewErrorValue("ERR - " + err.Error())
	return encodeError(e)
}

func SendNil() []byte {
	return []byte("$-1\r\n")
}

var store = make(map[string]Value)

func Ping() []byte {
	data, err := NewSimpleStringValue("PONG").Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Echo(v Value) []byte {
	data, err := v.Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Set(args []Value) []byte {
	var k, v Value
	var opt string
	var expiry int
	var err error

	if len(args) >= 2 {
		k = args[0]
		v = args[1]
	}

	if len(args) == 4 {
		opt = args[2].String()
		expiry, err = args[3].Integer()
		if err != nil {
			return SendError(err)
		}
	}

	var response []byte

	if old, ok := store[k.String()]; ok {
		response, err = NewBulkStringValue(old.String()).Encode()
		if err != nil {
			return SendError(err)
		}
	} else {
		response, err = NewSimpleStringValue("OK").Encode()
		if err != nil {
			return SendError(err)
		}
	}
	if v.Type() != BULK_STRING {
		v = NewBulkStringValue(v.String())
	}

	store[k.String()] = v

	if strings.ToUpper(opt) == "PX" && expiry > 0 {
		go func() {
			ch := time.After(time.Duration(expiry) * time.Millisecond)
			for {
				select {
				case <-ch:
					delete(store, k.String())
				}
			}
		}()
	}
	return response
}

func Get(k Value) []byte {
	if data, ok := store[k.String()]; ok {
		bytes, err := data.Encode()
		if err != nil {
			return SendError(err)
		}
		return bytes
	}
	return SendNil()
}
