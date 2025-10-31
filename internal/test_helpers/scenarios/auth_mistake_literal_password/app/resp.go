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

// RESPType represents the type of a RESP value
type RESPType int

const (
	TypeSimpleString RESPType = iota
	TypeError
	TypeInteger
	TypeBulkString
	TypeNil
	TypeArray
)

// RESPValue represents any RESP data type
type RESPValue struct {
	Type    RESPType
	Str     string
	Integer int
	Array   []RESPValue
}

type RESPCodec struct{}

func NewRESPCodec() *RESPCodec {
	return &RESPCodec{}
}

// Constructors for RESPValue
func SimpleString(s string) RESPValue {
	return RESPValue{Type: TypeSimpleString, Str: s}
}

func Error(s string) RESPValue {
	return RESPValue{Type: TypeError, Str: s}
}

func Integer(i int) RESPValue {
	return RESPValue{Type: TypeInteger, Integer: i}
}

func BulkString(s string) RESPValue {
	return RESPValue{Type: TypeBulkString, Str: s}
}

func Nil() RESPValue {
	return RESPValue{Type: TypeNil}
}

func Array(elements ...RESPValue) RESPValue {
	return RESPValue{Type: TypeArray, Array: elements}
}

// Encode converts a RESPValue to bytes
func (r *RESPCodec) Encode(value RESPValue) []byte {
	switch value.Type {
	case TypeSimpleString:
		return []byte(fmt.Sprintf("+%s\r\n", value.Str))
	case TypeError:
		return []byte(fmt.Sprintf("-%s\r\n", value.Str))
	case TypeInteger:
		return []byte(fmt.Sprintf(":%d\r\n", value.Integer))
	case TypeBulkString:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value.Str), value.Str))
	case TypeNil:
		return []byte("$-1\r\n")
	case TypeArray:
		var result []byte
		result = append(result, []byte(fmt.Sprintf("*%d\r\n", len(value.Array)))...)
		for _, element := range value.Array {
			result = append(result, r.Encode(element)...)
		}
		return result
	default:
		return []byte("-ERR unknown RESP type\r\n")
	}
}

// ReadCommand reads a RESP command and returns string arguments
func (r *RESPCodec) ReadCommand(reader *bufio.Reader) ([]string, error) {
	value, err := r.Read(reader)
	if err != nil {
		return nil, err
	}

	if value.Type != TypeArray {
		return nil, fmt.Errorf("expected array, got type %d", value.Type)
	}

	args := make([]string, len(value.Array))
	for i, elem := range value.Array {
		if elem.Type != TypeBulkString {
			return nil, fmt.Errorf("expected bulk string in array, got type %d", elem.Type)
		}
		args[i] = elem.Str
	}

	return args, nil
}

// Read reads any RESP value from the reader
func (r *RESPCodec) Read(reader *bufio.Reader) (RESPValue, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return RESPValue{}, err
	}

	switch firstByte {
	case RESP_SIMPLE_STRING:
		return r.readSimpleString(reader)
	case RESP_ERROR:
		return r.readError(reader)
	case RESP_INTEGER:
		return r.readInteger(reader)
	case RESP_BULK_STRING:
		return r.readBulkString(reader)
	case RESP_ARRAY:
		return r.readArray(reader)
	default:
		return RESPValue{}, fmt.Errorf("unknown RESP type: %c", firstByte)
	}
}

func (r *RESPCodec) readSimpleString(reader *bufio.Reader) (RESPValue, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	return SimpleString(strings.TrimSpace(line)), nil
}

func (r *RESPCodec) readError(reader *bufio.Reader) (RESPValue, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	return Error(strings.TrimSpace(line)), nil
}

func (r *RESPCodec) readInteger(reader *bufio.Reader) (RESPValue, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}
	i, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid integer: %s", line)
	}
	return Integer(i), nil
}

func (r *RESPCodec) readBulkString(reader *bufio.Reader) (RESPValue, error) {
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}

	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid bulk string length: %s", lengthStr)
	}

	// Handle nil bulk string
	if length == -1 {
		return Nil(), nil
	}

	// Read string content
	content := make([]byte, length)
	_, err = reader.Read(content)
	if err != nil {
		return RESPValue{}, err
	}

	// Read trailing \r\n
	_, err = reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}

	return BulkString(string(content)), nil
}

func (r *RESPCodec) readArray(reader *bufio.Reader) (RESPValue, error) {
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return RESPValue{}, err
	}

	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid array length: %s", lengthStr)
	}

	// Handle nil array
	if length == -1 {
		return Nil(), nil
	}

	// Read array elements
	elements := make([]RESPValue, length)
	for i := 0; i < length; i++ {
		elem, err := r.Read(reader)
		if err != nil {
			return RESPValue{}, err
		}
		elements[i] = elem
	}

	return Array(elements...), nil
}
