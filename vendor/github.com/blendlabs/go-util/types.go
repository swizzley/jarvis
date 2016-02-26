package util

import (
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
)

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type KVP struct {
	K string
	V interface{}
}

type KeyValuePairOfInt struct {
	Key   string
	Value int
}

type KVPI struct {
	K string
	V int
}

type KeyValuePairOfFloat struct {
	Key   string
	Value float64
}

type KVPF struct {
	K string
	V float64
}

type KeyValuePairOfString struct {
	Key   string
	Value string
}

type KVPS struct {
	K string
	V string
}

var BOOLEAN_TRUE Boolean = true
var BOOLEAN_FALSE Boolean = false

type Boolean bool

func (bit *Boolean) UnmarshalJSON(data []byte) error {
	as_string := strings.ToLower(string(data))
	if as_string == "1" || as_string == "true" {
		*bit = true
	} else if as_string == "0" || as_string == "false" {
		*bit = false
	} else if len(as_string) > 0 && (as_string[0] == '"' || as_string[0] == '\'') {
		cleaned := StripQuotes(as_string)
		return bit.UnmarshalJSON([]byte(cleaned))
	} else {
		return exception.Newf("Boolean unmarshal error: invalid input %s", as_string)
	}
	return nil
}

func (bit Boolean) AsBool() bool {
	if bit {
		return true
	} else {
		return false
	}
}

func OptionalUInt8(value uint8) *uint8 {
	return &value
}

func OptionalUInt16(value uint16) *uint16 {
	return &value
}

func OptionalUInt(value uint) *uint {
	return &value
}

func OptionalUInt64(value uint64) *uint64 {
	return &value
}

func OptionalInt16(value int16) *int16 {
	return &value
}

func OptionalInt(value int) *int {
	return &value
}

func OptionalInt64(value int64) *int64 {
	return &value
}

func OptionalFloat32(value float32) *float32 {
	return &value
}

func OptionalFloat64(value float64) *float64 {
	return &value
}

func OptionalString(value string) *string {
	return &value
}

func OptionalBool(value bool) *bool {
	return &value
}

func OptionalTime(value time.Time) *time.Time {
	return &value
}
