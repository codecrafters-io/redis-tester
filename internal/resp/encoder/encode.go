package resp_encoder

import (
	"fmt"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func Encode(v resp_value.Value) []byte {
	switch v.Type {
	case resp_value.INTEGER:
		return encodeInteger(v)
	case resp_value.SIMPLE_STRING:
		return encodeSimpleString(v)
	case resp_value.BULK_STRING:
		return encodeBulkString(v)
	case resp_value.ERROR:
		return encodeError(v)
	case resp_value.ARRAY:
		return encodeArray(v)
	case resp_value.NIL:
		return encodeNullBulkString()
	case resp_value.NIL_ARRAY:
		return encodeNullArray()
	default:
		panic(fmt.Sprintf("unsupported type: %v", v.Type))
	}
}

func EncodeFullResyncRDBFile(fileContents []byte) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s", len(fileContents), fileContents))
}

func encodeInteger(v resp_value.Value) []byte {
	int_value, err := strconv.Atoi(v.String())
	if err != nil {
		panic(err) // We only expect valid values to be passed in
	}

	return []byte(fmt.Sprintf(":%d\r\n", int_value))
}

func encodeSimpleString(v resp_value.Value) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", v.String()))
}

func encodeBulkString(v resp_value.Value) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Bytes()), v.Bytes()))
}

func encodeError(v resp_value.Value) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", v.String()))
}

func encodeArray(v resp_value.Value) []byte {
	res := []byte{}

	for _, elem := range v.Array() {
		res = append(res, Encode(elem)...)
	}

	return []byte(fmt.Sprintf("*%d\r\n%s", len(v.Array()), res))
}

func encodeNullBulkString() []byte {
	return []byte("$-1\r\n")
}

func encodeNullArray() []byte {
	return []byte("*-1\r\n")
}
