package modules

import (
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

func TestHandleConfigSet(t *testing.T) {
	assert := assert.New(t)

	c := &Config{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())
	handleErr := c.handleConfigSet(mb, core.MockMessage("config:foo bar"))
	assert.Nil(handleErr)
	assert.Equal("bar", mb.Configuration()["foo"])
}

func TestHandleConfigGet(t *testing.T) {
	assert := assert.New(t)
	c := &Config{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())
	mb.Configuration()["foo"] = "bar"

	gotMessage := ""
	mb.MockMessageHandler(func(b core.Bot, m *slack.Message) error {
		gotMessage = m.Text
		return nil
	})

	handleErr := c.handleConfigGet(mb, core.MockMessage("config:foo"))
	assert.Nil(handleErr)
	assert.NotEmpty(gotMessage)
	assert.True(strings.Contains(gotMessage, "foo"))
	assert.True(strings.Contains(gotMessage, "bar"))
}

func TestHandleConfig(t *testing.T) {
	assert := assert.New(t)
	c := &Config{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())
	mb.Configuration()["foo"] = "bar"

	handleErr := c.handleConfig(mb, core.MockMessage("config"))
	assert.Nil(handleErr)
}
