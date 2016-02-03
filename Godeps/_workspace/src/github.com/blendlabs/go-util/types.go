package util

import (
	"strings"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/blendlabs/go-exception"
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
