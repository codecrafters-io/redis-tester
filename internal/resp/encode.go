package resp

import (
	"fmt"
	"strconv"
)

func Encode(v Value) []byte {
	switch v.Type {
	case INTEGER:
		return encodeInteger(v)
	case SIMPLE_STRING:
		return encodeSimpleString(v)
	case BULK_STRING:
		return encodeBulkString(v)
	case ERROR:
		return encodeError(v)
	case ARRAY:
		return encodeArray(v)
	default:
		panic(fmt.Sprintf("unsupported type: %v", v.Type))
	}
}

func encodeInteger(v Value) []byte {
	int_value, err := strconv.Atoi(string(v.data))
	if err != nil {
		panic(err) // We only expect valid values to be passed in
	}

	return []byte(fmt.Sprintf(":%d\r\n", int_value))
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

func encodeArray(v Value) []byte {
	res := []byte{}

	for _, elem := range v.array {
		res = append(res, Encode(elem)...)
	}

	return []byte(fmt.Sprintf("*%d\r\n%s", len(v.array), res))
}
