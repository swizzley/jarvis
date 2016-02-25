package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestUnmarshalJSON(t *testing.T) {
	assert := assert.New(t)
	valid := []string{"1", "0", "true", "false", "True", "False", `"true"`, `"false"`}
	not_valid := []string{"foo", "123", "-1", "3.14", "", `""`}

	for index, value := range valid {
		var bit Boolean
		json_err := bit.UnmarshalJSON([]byte(value))
		assert.Nil(json_err)

		if index%2 == 0 {
			assert.True(bit.AsBool())
		} else {
			assert.False(bit.AsBool())
		}
	}

	for _, value := range not_valid {
		var bit Boolean
		json_err := bit.UnmarshalJSON([]byte(value))
		assert.NotNil(json_err)
	}
}
