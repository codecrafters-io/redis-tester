package resp_value

import (
	"encoding/json"
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/resp/formatter"
)

const (
	SIMPLE_STRING string = "simple string"
	INTEGER       string = "integer"
	BULK_STRING   string = "bulk string"
	ARRAY         string = "array"
	ERROR         string = "error"
	NIL           string = "null bulk string"
	NIL_ARRAY     string = "null array"
)

type Value struct {
	Type string

	// Each type might use a different field to store data
	bytes   []byte
	integer int
	array   []Value
}

func NewSimpleStringValue(s string) Value {
	return Value{
		Type:  SIMPLE_STRING,
		bytes: []byte(s),
	}
}

func NewBulkStringValue(s string) Value {
	return Value{
		Type:  BULK_STRING,
		bytes: []byte(s),
	}
}

func NewIntegerValue(i int) Value {
	return Value{
		Type:    INTEGER,
		integer: i,
	}
}

func NewErrorValue(err string) Value {
	return Value{
		Type:  ERROR,
		bytes: []byte(err),
	}
}

func NewStringArrayValue(strings []string) Value {
	values := make([]Value, len(strings))

	for i, s := range strings {
		values[i] = NewBulkStringValue(s)
	}

	return NewArrayValue(values)
}

func NewArrayValue(arr []Value) Value {
	return Value{
		Type:  ARRAY,
		array: arr,
	}
}

func NewNilValue() Value {
	return Value{
		Type: NIL,
	}
}

func NewNilArrayValue() Value {
	return Value{
		Type: NIL_ARRAY,
	}
}

func (v *Value) Bytes() []byte {
	return v.bytes
}

func (v *Value) Array() []Value {
	if v.Type == ARRAY {
		return v.array
	}

	return []Value{}
}

func (v *Value) String() string {
	return string(v.bytes)
}

func (v *Value) Integer() int {
	return v.integer
}

func (v *Value) Error() string {
	if v.Type == ERROR {
		return string(v.String())
	}
	return ""
}

func (v *Value) FormattedString() string {
	switch v.Type {
	case SIMPLE_STRING:
		return fmt.Sprintf("%q", v.String())
	case INTEGER:
		return fmt.Sprintf("%d", v.Integer())
	case BULK_STRING:
		return fmt.Sprintf("%q", v.String())
	case ARRAY:
		respJson, err := json.MarshalIndent(v.ToSerializable(), "", "  ")
		if err != nil {
			panic(fmt.Sprintf("Codecrafters Internal Error - Failed to encode to JSON: %#v", v.ToSerializable()))
		}
		return formatter.Prettify(respJson)
	case ERROR:
		return fmt.Sprintf("%q", v.String())
	case NIL:
		return "\"$-1\\r\\n\""
	case NIL_ARRAY:
		return "\"*-1\\r\\n\""
	}
	return ""
}

func (v Value) ToSerializable() interface{} {
	switch v.Type {
	case BULK_STRING:
		return v.String()
	case NIL:
		return "$-1\r\n"
	case NIL_ARRAY:
		return "*-1\r\n"
	case INTEGER:
		return v.Integer()
	case ARRAY:
		arr := v.Array()
		result := make([]interface{}, len(arr))
		for i, elem := range arr {
			result[i] = elem.ToSerializable()
		}
		return result
	default:
		return v.String()
	}
}
