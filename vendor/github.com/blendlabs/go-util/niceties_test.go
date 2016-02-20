package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestTernary(t *testing.T) {
	assert := assert.New(t)

	foo := Ternary(true, "foo", "bar").(string)
	assert.Equal("foo", foo)
	bar := Ternary(false, "foo", "bar").(string)
	assert.Equal("bar", bar)
}
