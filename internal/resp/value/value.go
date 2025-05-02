package resp_value

import (
	"fmt"
	"strings"
)

const (
	SIMPLE_STRING string = "SIMPLE_STRING"
	INTEGER       string = "INTEGER"
	BULK_STRING   string = "BULK_STRING"
	ARRAY         string = "ARRAY"
	ERROR         string = "ERROR"
	NIL           string = "NIL"
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
		formattedStrings := make([]string, len(v.Array()))

		for i, value := range v.Array() {
			formattedStrings[i] = value.FormattedString()
		}

		return fmt.Sprintf("[%v]", strings.Join(formattedStrings, ", "))
	case ERROR:
		return fmt.Sprintf("%q", "ERR: "+v.String())
	case NIL:
		return "\"$-1\\r\\n\""
	}

	return ""
}

func (v Value) ToSerializable() interface{} {
	switch v.Type {
	case BULK_STRING:
		return v.String()
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
