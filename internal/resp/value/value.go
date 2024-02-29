package resp_value

import (
	"fmt"
	"strconv"
)

const (
	SIMPLE_STRING string = "SIMPLE_STRING"
	INTEGER       string = "INTEGER"
	BULK_STRING   string = "BULK_STRING"
	ARRAY         string = "ARRAY"
	ERROR         string = "ERROR"
)

type Value struct {
	Type  string
	data  []byte
	array []Value
}

func NewSimpleStringValue(s string) Value {
	return Value{
		Type: SIMPLE_STRING,
		data: []byte(s),
	}
}

func NewBulkStringValue(s string) Value {
	return Value{
		Type: BULK_STRING,
		data: []byte(s),
	}
}

func NewIntegerValue(i int) Value {
	return Value{
		Type: INTEGER,
		data: []byte(fmt.Sprint(i)),
	}
}

func NewErrorValue(err string) Value {
	return Value{
		Type: ERROR,
		data: []byte(err),
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

func (v *Value) Bytes() []byte {
	return v.data
}

func (v *Value) Array() []Value {
	if v.Type == ARRAY {
		return v.array
	}

	return []Value{}
}

func (v *Value) String() string {
	return string(v.data)
}

func (v *Value) Integer() int {
	i, err := strconv.Atoi(string(v.data))
	if err != nil {
		panic("invalid integer") // This should never happen since we create values
	}
	return i
}

func (v *Value) FormattedString() string {
	switch v.Type {
	case SIMPLE_STRING:
		return fmt.Sprintf("%q", v.String())
	case INTEGER:
		return fmt.Sprintf("%q", v.Integer())
	case BULK_STRING:
		return fmt.Sprintf("%q", v.String())
	case ARRAY:
		return fmt.Sprintf("%q", v.Array())
	case ERROR:
		return fmt.Sprintf("%q", "ERR: "+v.String())
	}

	return ""
}
