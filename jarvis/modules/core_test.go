package modules

import (
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

func TestHandleHelp(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())

	err := c.handleHelp(mb, core.MockMessage("help"))
	assert.Nil(err)
}

func TestHandleTell(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())

	gotMessage := ""
	mb.MockMessageHandler(func(b core.Bot, m *slack.Message) error {
		gotMessage = m.Text
		return nil
	})

	err := c.handleTell(mb, core.MockMessage("tell <@TESTUSER> they're cool"))
	assert.Nil(err)
	assert.Equal("<@TESTUSER> you are cool", gotMessage)
}

func TestHandleChannels(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())
	err := c.handleChannels(mb, core.MockMessage("channels"))
	assert.Nil(err)
}

func TestHandleMentionCatchAllSalutation(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())

	gotMessage := ""
	mb.MockMessageHandler(func(b core.Bot, m *slack.Message) error {
		gotMessage = m.Text
		return nil
	})

	message := "hey <@BOT>"
	assert.True(core.IsSalutation(message))
	err := c.handleSalutation(mb, core.MockMessage(message))
	assert.Nil(err)
	assert.False(strings.Contains(gotMessage, "how to respond"))
}

func TestHandleMentionCatchNonSalutation(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())

	gotMessage := ""
	mb.MockMessageHandler(func(b core.Bot, m *slack.Message) error {
		gotMessage = m.Text
		return nil
	})

	message := "this is a test message"
	assert.False(core.IsSalutation(message))
	err := c.handleMentionCatchAll(mb, core.MockMessage(message))
	assert.Nil(err)
	println(gotMessage)
	assert.True(strings.Contains(gotMessage, "how to respond"))
}
