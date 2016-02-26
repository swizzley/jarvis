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

	handleErr := c.handleHelp(mb, core.MockMessage("help"))
	assert.Nil(handleErr)
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

	handleErr := c.handleTell(mb, core.MockMessage("tell <@TESTUSER> they're cool"))
	assert.Nil(handleErr)
	assert.Equal("<@TESTUSER> you are cool", gotMessage)
}

func TestHandleChannels(t *testing.T) {
	assert := assert.New(t)
	c := &Core{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())
	handleErr := c.handleChannels(mb, core.MockMessage("channels"))
	assert.Nil(handleErr)
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
	handleErr := c.handleSalutation(mb, core.MockMessage(message))
	assert.Nil(handleErr)
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
	handleErr := c.handleMentionCatchAll(mb, core.MockMessage(message))
	assert.Nil(handleErr)
	println(gotMessage)
	assert.True(strings.Contains(gotMessage, "how to respond"))
}
