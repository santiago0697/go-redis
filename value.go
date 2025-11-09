package main

import (
	"bytes"
	"strconv"
	"strings"
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

func (v *Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshallArray()
	case "bulk":
		return v.marshallBulk()
	case "string":
		return v.marshallString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

func (v *Value) marshallError() []byte {
	var buf bytes.Buffer
	buf.Grow(len(v.str) + 3)

	buf.WriteByte(ERROR)
	buf.WriteString(v.str)
	buf.WriteString("\r\n")

	return buf.Bytes()
}

func (v *Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

func ErrValue(str string) Value {
	return Value{typ: "error", str: str}
}

func OkValue() Value {
	return Value{typ: "string", str: "OK"}
}

func NullValue() Value {
	return Value{str: "null"}
}

func BulkValue(bulk string) Value {
	return Value{typ: "bulk", bulk: bulk}
}

func ArrayValue (values []Value) Value {
	return Value{typ: "array", array: values}
}

func (v *Value) marshallString() []byte {
	var buf bytes.Buffer
	buf.Grow(len(v.str) + 3)

	buf.WriteByte(STRING)
	buf.WriteString(v.str)
	buf.WriteString("\r\n")

	return buf.Bytes()
}

func (v *Value) marshallBulk() []byte {
	var buf bytes.Buffer
	buf.Grow(len(v.bulk) + 3)

	buf.WriteByte(BULK)
	buf.WriteString(strconv.Itoa(len(v.bulk)))
	buf.WriteString("\r\n")
	buf.WriteString(v.bulk)
	buf.WriteString("\r\n")

	return buf.Bytes()
}

func (v *Value) marshallArray() []byte {
	var buf bytes.Buffer

	buf.WriteByte(ARRAY)
	buf.WriteString(strconv.Itoa(len(v.array)))
	buf.WriteString("\r\n")

	for _,val := range v.array {
		buf.Write(val.Marshal())
	}

	return buf.Bytes()
}

func (v *Value) String() string {
	var buf strings.Builder

	buf.WriteString(v.bulk)

	for _, val := range v.array {
		buf.WriteString(val.String())
		buf.WriteByte(' ')
	}

	return buf.String()
}
