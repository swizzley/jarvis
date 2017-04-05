package core

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestChannelRegistry(t *testing.T) {
	assert := assert.New(t)

	cr := ChannelRegistry{}
	cr.Register("foo", "bar", "baz")
	assert.NotEmpty(cr.UsersInChannel("foo", "bar"))
}
